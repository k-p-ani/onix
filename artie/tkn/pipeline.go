package tkn

type Pipeline struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

type TaskRef struct {
	Name string `yaml:"name"`
}

type Tasks struct {
	Name      string    `yaml:"name"`
	TaskRef   TaskRef   `yaml:"taskRef"`
	Resources Resources `yaml:"resources"`
}