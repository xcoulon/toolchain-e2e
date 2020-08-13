package perf

import (
	"testing"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/pkg/apis/toolchain/v1alpha1"
	. "github.com/codeready-toolchain/toolchain-e2e/testsupport"
	. "github.com/codeready-toolchain/toolchain-e2e/wait"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestPerformances(t *testing.T) {
	// given
	ctx, awaitility := WaitForDeployments(t, &toolchainv1alpha1.UserSignupList{})
	// defer ctx.Cleanup()

	// host and member cluster statuses should be available at this point
	t.Run("verify cluster statuses are valid", func(t *testing.T) {
		t.Run("verify member cluster status", func(t *testing.T) {
			VerifyMemberStatus(t, awaitility.Member())
		})

		t.Run("verify overall toolchain status", func(t *testing.T) {
			VerifyToolchainStatus(t, awaitility.Host())
		})
	})

	// host metrics should be available at this point
	t.Run("verify metrics servers", func(t *testing.T) {
		t.Run("verify host metrics server", func(t *testing.T) {
			VerifyHostMetricsService(t, awaitility.Host())
		})
	})

	t.Run("10 users", func(t *testing.T) {
		// Create multiple accounts and let them get provisioned while we are executing the main flow for "johnsmith" and "extrajohn"
		// We will verify them in the end of the test
		users := CreateMultipleSignups(t, ctx, awaitility, 10)
		for _, user := range users {
			awaitility.Host().WaitForMasterUserRecord(user.Spec.Username, UntilMasterUserRecordHasCondition(Provisioned()))
		}
		// when deleting the host-operator pod
		err := awaitility.Host().DeletePods(client.MatchingLabels{"name": "host-operator"})
		require.NoError(t, err)
		// then measure time it takes to have an empty queue on the master-user-records
		awaitility.Host().WaitUntilMetricsCounterHasValue("controller_runtime_reconcile_total", "controller", "usersignup-controller", 10)
		awaitility.Host().WaitUntilMetricsCounterHasValue("workqueue_depth", "name", "usersignup-controller", 0)
	})

}
