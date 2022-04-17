package k8s

import (
	"context"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Secret struct {
	Id       string
	Password string
}

func GetSecret() (Secret, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return Secret{}, err
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return Secret{}, err
	}

	const (
		namespace = "rabbitmq"
		name      = "rabbitmq-default-user"
	)

	secret, err := clientSet.CoreV1().Secrets(namespace).Get(context.Background(), name, metaV1.GetOptions{})
	if err != nil {
		return Secret{}, err
	}
	return Secret{
		Id:       string(secret.Data["username"]),
		Password: string(secret.Data["password"]),
	}, nil
}
