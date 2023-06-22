package targetcluster

import (
	"context"
	"fmt"
)

func (tc *TargetCluster) GetOperatorCapabilities(ctx context.Context, namespace string, operatorName string) (map[string]string, error) {
	configMapName := fmt.Sprintf("pds-%s-operator-capabilities", operatorName)
	configMap, err := tc.GetConfigMap(ctx, namespace, configMapName)
	if err != nil {
		return nil, err
	}
	return configMap.Data, nil
}
