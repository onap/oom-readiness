package client

import (
	"context"
	"log"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r ReadinessClient) isDaemonSetReady(namespace string, name string) bool {
	ds, err := r.Client.AppsV1().DaemonSets(namespace).Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		log.Printf("Error while getting DeamonSet %s: %v", name, err)
		return false
	}
	if ds.Status.DesiredNumberScheduled == ds.Status.NumberReady {
		log.Printf("DaemonSet: %d/%d nodes ready --> %s is ready", ds.Status.NumberReady, ds.Status.DesiredNumberScheduled, ds.Name)
		return true
	}
	return false
}
