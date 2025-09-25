package client

import (
	"context"
	"log"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r ReadinessClient) IsDeploymentReady(namespace string, name string) bool {
	deployment, err := r.Client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		log.Printf("Error during get of deployment %s: %v", name, err)
	}
	if isDeploymentReady(*deployment) {
		log.Printf("Deployment %s is ready", name)
		return true
	} else {
		log.Printf("Deployment %s is NOT ready", name)
		return false
	}
}

func isDeploymentReady(dpl appsv1.Deployment) bool {
	return dpl.Status.UnavailableReplicas == 0 &&
		(dpl.Status.UpdatedReplicas == 0 || dpl.Status.UpdatedReplicas == *dpl.Spec.Replicas) &&
		dpl.Status.Replicas == *dpl.Spec.Replicas &&
		dpl.Status.ObservedGeneration == dpl.Generation
}
