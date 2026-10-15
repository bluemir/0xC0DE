package auth

import (
	"context"

	"gopkg.in/yaml.v3"

	_ "embed"

	"github.com/sirupsen/logrus"
)

//go:embed policy.yaml
var policyData []byte

func (m *Manager) initialize(ctx context.Context) error {
	data := &struct {
		Roles    []Role
		Bindings []struct {
			Subject Subject
			Role    string
		}
		Groups []string
	}{}
	if err := yaml.Unmarshal(policyData, data); err != nil {
		return err
	}

	{
		buf, _ := yaml.Marshal(data)
		logrus.Tracef("%s", string(buf))
	}

	for _, group := range data.Groups {
		if _, err := m.EnsureGroup(ctx, group); err != nil {
			logrus.Warn(err)
		}
	}

	for _, role := range data.Roles {
		if _, err := m.CreateRole(ctx, role.Name, role.Rules); err != nil {
			logrus.Warn(err)
		}
	}

	for _, binding := range data.Bindings {
		if err := m.AssignRole(ctx, binding.Subject, binding.Role); err != nil {
			logrus.Warn(err)
		} else {
			logrus.Info("add binding", binding.Subject, binding.Role)
		}
	}

	return nil
}
