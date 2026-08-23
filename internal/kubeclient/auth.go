package kubeclient

import (
	"context"
	"fmt"
	"log/slog"

	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *KubeClient) doSSAR(ctx context.Context, namespace, verb, group, version, resource, name string) error {
  client := k.kubernetes
  ssar := &authv1.SelfSubjectAccessReview{
    Spec: authv1.SelfSubjectAccessReviewSpec{
      ResourceAttributes: &authv1.ResourceAttributes{
        Namespace: namespace,
        Verb: verb,
        Group: group,
        Version: version,
        Resource: resource,
        Name: name,
      },
    },
  }
  res, err := client.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, ssar, metav1.CreateOptions{})
  if err != nil {
    return fmt.Errorf("check ssar for %s %s/%s: %w", verb, version, resource, err)
  }
  if !res.Status.Allowed {
    return fmt.Errorf("not allowed to %s %s/%s", verb, version, resource)
  }
  return nil
}

// Tests the Kubernetes client for access to GatewayAPI resources.
func (k *KubeClient) CheckPermissions(ctx context.Context) error {
  slog.Info("checking kube client permissions")
  if err := k.doSSAR(ctx, "", "list", "gateway.networking.k8s.io", "v1", "httproutes", ""); err != nil {
    return fmt.Errorf("check permissions: %v", err)
  }
  slog.Info("permissions okay")
  return nil
}
