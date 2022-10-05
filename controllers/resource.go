package controllers

import (
	"github.com/Sion-L/operator-demo/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//v1 "k8s.io/client-go/applyconfigurations/apps/v1"
)

func MutateService(app *v1beta1.MyApp, svc *corev1.Service) {
	// TODO
	svc.Spec = corev1.ServiceSpec{
		ClusterIP: svc.Spec.ClusterIP,
		Ports:     app.Spec.Ports,
		Type:      corev1.ServiceTypeNodePort,
		Selector: map[string]string{
			"myapp": app.Name,
		},
	}
}

func newContainers(app *v1beta1.MyApp) []corev1.Container {
	// 从crd结构中的ports列表取端口
	containerPorts := []corev1.ContainerPort{}
	for _, svcPort := range app.Spec.Ports {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			ContainerPort: svcPort.TargetPort.IntVal,
		})
	}
	return []corev1.Container{
		{
			Name:      app.Name,
			Image:     app.Spec.Image,
			Resources: app.Spec.Resources,
			Env:       app.Spec.Envs,
			Ports:     containerPorts,
		},
	}
}

func MutateDeployment(app *v1beta1.MyApp, deploy *appsv1.Deployment) {
	labels := map[string]string{"myapp": app.Name}
	selector := &metav1.LabelSelector{
		MatchLabels: labels,
	}
	deploy.Spec = appsv1.DeploymentSpec{
		Replicas: app.Spec.Size,
		Template: corev1.PodTemplateSpec{ // podTemplate的信息
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				Containers: newContainers(app),
			},
		},
		Selector: selector,
	}
}
