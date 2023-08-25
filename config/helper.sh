TESTS=${TESTS:-"iam,namespace,backup,restore,deployment,portworxcsi,reporting,capabilities"}
TEST_IMAGE_NAME=${TEST_IMAGE_NAME:-"docker.io/portworx/pds-integration-test"}
TEST_IMAGE_VERSION=${TEST_IMAGE_VERSION:-"master"}
ENVIRONMENT_FILE=${ENVIRONMENT_FILE:-"environment.template"}
TEST_NAMESPACE=${TEST_NAMESPACE:-"pds-integration-tests"}

TEST_SLEEP_SECONDS=${TEST_SLEEP_SECONDS:-60}
TEST_TIMEOUT_SECONDS=${TEST_TIMEOUT_SECONDS:-3600}
LOGS_PATH=${LOGS_PATH:-"./logs"}

KUSTOMIZATION_TEMPLATE="
namespace: pds-integration-tests
namePrefix: pds-

resources:
- ./base/namespace.yml
- ./base/rbac.yaml
%b

images:
  - name: pdstestimage
    newName: %s
    newTag: %s

configMapGenerator:
- name: config
  envs:
  - %s
- name: helm-repository
  files:
    - repositories.yml
- name: dataservices
  files:
    - dataservices.yml
"

function help(){
  printf 'Description : This script generates the kustomization files

Available Commands:
help - helper.sh help
config - <vars> helper.sh config
  Eg: TESTS="iam" helper.sh config

Available variables:
TESTS - Comma separated test names
TEST_IMAGE_NAME - Test image name
TEST_IMAGE_VERSION - Test image version
ENVIRONMENT_FILE - File path with environment variables for tests'
}

function config(){
  RESOURCES=""
  for i in $(echo $TESTS | sed "s/,/ /g")
  do
      RESOURCES=$RESOURCES"- ./base/"$i"_test.yml\n"
  done

  printf "$KUSTOMIZATION_TEMPLATE" "$RESOURCES" "$TEST_IMAGE_NAME" "$TEST_IMAGE_VERSION" "$ENVIRONMENT_FILE"
}

function registrationConfig(){
  RESOURCES="- ./base/register_test.yml\n"

  printf "$KUSTOMIZATION_TEMPLATE" "$RESOURCES" "$TEST_IMAGE_NAME" "$TEST_IMAGE_VERSION" "$ENVIRONMENT_FILE"
}

function deregistrationConfig(){
  RESOURCES="- ./base/deregister_test.yml\n"

  printf "$KUSTOMIZATION_TEMPLATE" "$RESOURCES" "$TEST_IMAGE_NAME" "$TEST_IMAGE_VERSION" "$ENVIRONMENT_FILE"
}

function getJobCompletionStatus(){
  if kubectl wait --for=condition=complete --timeout=0 -n "$1" job/"$2" 2>/dev/null; then
    printf "[Success] job %s/%s completed\n" "$1" "$2"
    return 0
  fi

  if kubectl wait --for=condition=failed --timeout=0 -n "$1" job/"$2" 2>/dev/null; then
    printf "[Failure] job %s/%s failed\n" "$1" "$2"
    return 1
  fi

  printf "[Running] job %s/%s is still running\n" "$1" "$2"
  return 2
}

function getJobsCompletionStatus(){
  allCompleted=true
  read -ra jobs <<< $(kubectl -n="$TEST_NAMESPACE" get jobs -ojsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}')

  for job in "${jobs[@]}";
  do
    printf "checking for Job %s/%s\n" "$TEST_NAMESPACE" "$job"

    getJobCompletionStatus "$TEST_NAMESPACE" "$job"
    if [[ $? -eq 2 ]]; then
      allCompleted=false
    fi
  done

  if [[ $allCompleted == false ]]; then
    printf "jobs are still running\n"
    return 1
  fi

  printf "all jobs have completed successfully\n"
  return 0
}

function evaluateResult(){
  allPassed=true
  read -ra jobs <<< $(kubectl -n="$TEST_NAMESPACE" get jobs -ojsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}')

  for job in "${jobs[@]}";
  do
    printf "checking for Job %s/%s\n" "$TEST_NAMESPACE" "$job"

    if ! getJobCompletionStatus "$TEST_NAMESPACE" "$job"; then
      printf "[Failure] job %s/%s failed" "$TEST_NAMESPACE" "$job"
      allPassed=false
    fi
  done

  if [[ $allPassed == false ]]; then
    return 1
  fi

  printf "[Success] all jobs have completed successfully\n"
  return 0
}

function waitForTestsCompletion(){
  printf "waiting for Test Suites in namespace %s\n" "$TEST_NAMESPACE"

  CURRENT_ITERATION=0
  TEST_ITERATIONS=$((TEST_TIMEOUT_SECONDS/TEST_SLEEP_SECONDS))

  while true; do
    if [[ $CURRENT_ITERATION -ge $TEST_ITERATIONS ]]; then
      printf "[TIMEOUT] waiting for job completion timed out after %s iterations\n" $CURRENT_ITERATION
      return 1
    fi

    if getJobsCompletionStatus "$TEST_NAMESPACE"; then
      return 0
    fi

    CURRENT_ITERATION=$((CURRENT_ITERATION + 1))

    printf "sleeping for %s seconds\n" "$TEST_SLEEP_SECONDS"
    sleep "$TEST_SLEEP_SECONDS"
  done
}

function captureLogs(){
  mkdir -p "$LOGS_PATH"

  read -ra pods <<< $(kubectl -n="$TEST_NAMESPACE" get pods -ojsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}')

  for pod in "${pods[@]}";
  do
    LOG_FILE_NAME="$LOGS_PATH/$pod.log"

    printf "collecting logs for Job %s/%s at %s\n" "$TEST_NAMESPACE" "$pod" "$LOG_FILE_NAME"

    if ! $(kubectl logs -n="$TEST_NAMESPACE" pods/"$pod" --container='tests' --prefix=true > "$LOG_FILE_NAME"); then
      printf "[Failure] collecting logs for %s/%s failed" "$TEST_NAMESPACE" "$pod"
    fi
  done

  printf "logs have been collected at %s/" "$LOGS_PATH"
}

"$@"



