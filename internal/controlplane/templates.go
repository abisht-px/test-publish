package controlplane

// Info for a single template.
// TODO 4915: Unexport definitions when all referencing helpers are attached to the controlPlane.
type TemplateInfo struct {
	ID   string
	Name string
}

// Info for all app config and resource templates which belong to a data service.
// TODO 4915: Unexport definitions when all referencing helpers are attached to the controlPlane.
type DataServiceTemplateInfo struct {
	AppConfigTemplates []TemplateInfo
	ResourceTemplates  []TemplateInfo
}
