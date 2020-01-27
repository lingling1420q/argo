package e2e

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo/test/e2e/fixtures"
)

type CLISuite struct {
	fixtures.E2ESuite
}

func (s *CLISuite) BeforeTest(suiteName, testName string) {
	s.E2ESuite.BeforeTest(suiteName, testName)

}

func (s *CLISuite) AfterTest(suiteName, testName string) {
	s.E2ESuite.AfterTest(suiteName, testName)
}

func (s *CLISuite) TestCompletion() {
	s.Given().RunCli([]string{"completion", "bash"}, func(t *testing.T, output string, err error) {
		assert.NoError(t, err)
		assert.Contains(t, output, "bash completion for argo")
	})
}

func (s *CLISuite) TestToken() {
	s.Given().RunCli([]string{"token"}, func(t *testing.T, output string, err error) {
		assert.NoError(t, err)
		token, err := s.GetServiceAccountToken()
		assert.NoError(t, err)
		assert.Equal(t, token, output)
	})
}

func (s *CLISuite) TestRoot() {
	s.Given().RunCli([]string{"submit", "smoke/basic.yaml"}, func(t *testing.T, output string, err error) {
		assert.NoError(t, err)
		assert.Contains(t, output, "Namespace:")
		assert.Contains(t, output, "ServiceAccount:")
		assert.Contains(t, output, "Status:")
		assert.Contains(t, output, "Created:")
	})
	s.Given().RunCli([]string{"list"}, func(t *testing.T, output string, err error) {
		assert.NoError(t, err)
		assert.Contains(t, output, "NAME")
		assert.Contains(t, output, "STATUS")
		assert.Contains(t, output, "AGE")
		assert.Contains(t, output, "DURATION")
		assert.Contains(t, output, "PRIORITY")
	})
	s.Given().RunCli([]string{"get", "basic"}, func(t *testing.T, output string, err error) {
		assert.NoError(t, err)
		assert.Contains(t, output, "Namespace:")
		assert.Contains(t, output, "ServiceAccount:")
		assert.Contains(t, output, "Status:")
		assert.Contains(t, output, "Created:")
	})

	var createdWorkflowName string
	s.Given().CronWorkflow("@testdata/basic.yaml").
		When().
		CreateCronWorkflow().
		RunCli([]string{"submit", "--from", "CronWorkflow/test-cron-wf-basic"}, func(t *testing.T, output string, err error) {
			assert.NoError(t, err)
			assert.Contains(t, output, "Name:                test-cron-wf-basic-")
			r := regexp.MustCompile(`Name:\s+?(test-cron-wf-basic-[a-z0-9]+)`)
			res := r.FindStringSubmatch(output)
			if len(res) != 2 {
				assert.Fail(t, "Internal test error, please report a bug")
			}
			createdWorkflowName = res[1]
		}).
		WaitForWorkflowName(createdWorkflowName, 15*time.Second).
		Then().
		ExpectWorkflowName(createdWorkflowName, func(t *testing.T, metadata *metav1.ObjectMeta, status *wfv1.WorkflowStatus) {
			assert.Equal(t, wfv1.NodeSucceeded, status.Phase)
		})

	s.Given().WorkflowTemplate("@smoke/workflow-template-whalesay-template.yaml").
		When().
		CreateWorkflowTemplates().
		RunCli([]string{"submit", "--from", "WorkflowTemplate/workflow-template-whalesay-template"}, func(t *testing.T, output string, err error) {
			assert.Errorf(t, err, "When submitting a Workflow from a WorkflowTemplate an entrypoint must be passed with --entrypoint")
		}).
		RunCli([]string{"submit", "--from", "WorkflowTemplate/workflow-template-whalesay-template", "--entrypoint", "whalesay-template", "--parameter", "message=TEST"}, func(t *testing.T, output string, err error) {
			assert.NoError(t, err)
			assert.Contains(t, output, "Name:                workflow-template-whalesay-template-")
			r := regexp.MustCompile(`Name:\s+?(workflow-template-whalesay-template-[a-z0-9]+)`)
			res := r.FindStringSubmatch(output)
			if len(res) != 2 {
				assert.Fail(t, "Internal test error, please report a bug")
			}
			createdWorkflowName = res[1]
		}).
		WaitForWorkflowName(createdWorkflowName, 15*time.Second).
		Then().
		ExpectWorkflowName(createdWorkflowName, func(t *testing.T, metadata *metav1.ObjectMeta, status *wfv1.WorkflowStatus) {
			assert.Equal(t, wfv1.NodeSucceeded, status.Phase)
		})
}

func (s *CLISuite) TestTemplate() {

	s.Given().RunCli([]string{"template", "lint", "smoke/workflow-template-whalesay-template.yaml"}, func(t *testing.T, output string, err error) {
		assert.NoError(t, err)
		assert.Contains(t, output, "validated")
	})

	s.Given().RunCli([]string{"template", "create", "smoke/workflow-template-whalesay-template.yaml"}, func(t *testing.T, output string, err error) {
		assert.NoError(t, err)
		assert.Contains(t, output, "Name:")
		assert.Contains(t, output, "Namespace:")
		assert.Contains(t, output, "Created:")
	})

	s.Given().RunCli([]string{"template", "list"}, func(t *testing.T, output string, err error) {
		assert.NoError(t, err)
		assert.Contains(t, output, "NAME")
	})

	s.Given().RunCli([]string{"template", "get", "not-found"}, func(t *testing.T, output string, err error) {
		assert.Error(t, err, "exit status 1")
		assert.Contains(t, output, `"not-found" not found`)
	}).RunCli([]string{"template", "get", "workflow-template-whalesay-template"}, func(t *testing.T, output string, err error) {
		if assert.NoError(t, err) {
			assert.Contains(t, output, "Name:")
			assert.Contains(t, output, "Namespace:")
			assert.Contains(t, output, "Created:")
		}
	})

	s.Given().RunCli([]string{"template", "delete", "workflow-template-whalesay-template"}, func(t *testing.T, output string, err error) {
		assert.NoError(t, err)
	})
}

func (s *CLISuite) TestCron() {

	s.Given().RunCli([]string{"cron", "create", "testdata/basic.yaml"}, func(t *testing.T, output string, err error) {
		assert.NoError(t, err)
		assert.Contains(t, output, "Name:")
		assert.Contains(t, output, "Namespace:")
		assert.Contains(t, output, "Created:")
		assert.Contains(t, output, "Schedule:")
		assert.Contains(t, output, "Suspended:")
		assert.Contains(t, output, "StartingDeadlineSeconds:")
		assert.Contains(t, output, "ConcurrencyPolicy:")
	})

	s.Given().RunCli([]string{"cron", "list"}, func(t *testing.T, output string, err error) {
		assert.NoError(t, err)
		assert.Contains(t, output, "NAME")
		assert.Contains(t, output, "AGE")
		assert.Contains(t, output, "LAST RUN")
		assert.Contains(t, output, "SCHEDULE")
		assert.Contains(t, output, "SUSPENDED")
	})

	s.Given().RunCli([]string{"cron", "get", "not-found"}, func(t *testing.T, output string, err error) {
		assert.Error(t, err, "exit status 1")
		assert.Contains(t, output, `"not-found" not found`)
	}).RunCli([]string{"cron", "get", "test-cron-wf-basic"}, func(t *testing.T, output string, err error) {
		if assert.NoError(t, err) {
			assert.Contains(t, output, "Name:")
			assert.Contains(t, output, "Namespace:")
			assert.Contains(t, output, "Created:")
			assert.Contains(t, output, "Schedule:")
			assert.Contains(t, output, "Suspended:")
			assert.Contains(t, output, "StartingDeadlineSeconds:")
			assert.Contains(t, output, "ConcurrencyPolicy:")
		}
	})

	s.Given().RunCli([]string{"cron", "delete", "test-cron-wf-basic"}, func(t *testing.T, output string, err error) {
		assert.NoError(t, err)
	})
}

func (s *CLISuite) TestArchive() {
	if !s.Persistence.IsEnabled() {
		s.T().SkipNow()
	}
	var uid types.UID
	s.Given().
		Workflow("@smoke/basic.yaml").
		When().
		SubmitWorkflow().
		WaitForWorkflow(30*time.Second).
		Then().
		ExpectWorkflow(func(t *testing.T, metadata *metav1.ObjectMeta, status *wfv1.WorkflowStatus) {
			uid = metadata.UID
		}).
		RunCli([]string{"archive", "list"}, func(t *testing.T, output string, err error) {
			if assert.NoError(t, err) {
				assert.Contains(t, output, "NAMESPACE NAME")
				assert.Contains(t, output, "argo basic")
			}
		}).
		RunCli([]string{"archive", "get", string(uid)}, func(t *testing.T, output string, err error) {
			if assert.NoError(t, err) {
				assert.Contains(t, output, "Succeeded")
			}
		}).
		RunCli([]string{"archive", "delete", string(uid)}, func(t *testing.T, output string, err error) {
			if assert.NoError(t, err) {
				assert.Contains(t, output, "Archived workflow")
				assert.Contains(t, output, "deleted")
			}
		})
}

func TestCliSuite(t *testing.T) {
	suite.Run(t, new(CLISuite))
}
