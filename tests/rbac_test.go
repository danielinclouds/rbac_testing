package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/matryer/is"
	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var clientset *kubernetes.Clientset
var blueUser = "dave@gmail.com"
var blueNamespace = "app1"
var redNamespace = "app4"

func TestRBAC(t *testing.T) {

	before(t)
	is := is.New(t)

	t.Run("Developer can create deployment", func(t *testing.T) {

		sar := &authv1.SubjectAccessReview{
			Spec: authv1.SubjectAccessReviewSpec{
				ResourceAttributes: &authv1.ResourceAttributes{
					Namespace: blueNamespace,
					Verb:      "create",
					Group:     "apps",
					Resource:  "deployments",
				},
				User: blueUser,
			},
		}

		sar, err := clientset.AuthorizationV1().
			SubjectAccessReviews().
			Create(context.TODO(), sar, metav1.CreateOptions{})

		is.NoErr(err)
		is.True(sar.Status.Allowed)

	})

	t.Run("Developer can't create role binding", func(t *testing.T) {

		sar := &authv1.SubjectAccessReview{
			Spec: authv1.SubjectAccessReviewSpec{
				ResourceAttributes: &authv1.ResourceAttributes{
					Namespace: blueNamespace,
					Verb:      "create",
					Group:     "rbac.authorization.k8s.io",
					Resource:  "rolebindings",
				},
				User: blueUser,
			},
		}

		sar, err := clientset.AuthorizationV1().
			SubjectAccessReviews().
			Create(context.TODO(), sar, metav1.CreateOptions{})

		is.NoErr(err)
		is.True(!sar.Status.Allowed)

	})

	t.Run("Developers can't change namespace labels", func(t *testing.T) {

		sar := &authv1.SubjectAccessReview{
			Spec: authv1.SubjectAccessReviewSpec{
				ResourceAttributes: &authv1.ResourceAttributes{
					Namespace: "",
					Verb:      "label",
					Resource:  "namespaces",
				},
				User: blueUser,
			},
		}

		sar, err := clientset.AuthorizationV1().
			SubjectAccessReviews().
			Create(context.TODO(), sar, metav1.CreateOptions{})

		is.NoErr(err)
		is.True(!sar.Status.Allowed)

	})

	t.Run("User from team blue can't list pods in team red namespace", func(t *testing.T) {

		sar := &authv1.SubjectAccessReview{
			Spec: authv1.SubjectAccessReviewSpec{
				ResourceAttributes: &authv1.ResourceAttributes{
					Namespace: redNamespace,
					Verb:      "get",
					Resource:  "pods",
				},
				User: blueUser,
			},
		}

		sar, err := clientset.AuthorizationV1().
			SubjectAccessReviews().
			Create(context.TODO(), sar, metav1.CreateOptions{})

		is.NoErr(err)
		is.True(!sar.Status.Allowed)

	})

	t.Run("Developer can't list pods in kube-system namespace", func(t *testing.T) {

		sar := &authv1.SubjectAccessReview{
			Spec: authv1.SubjectAccessReviewSpec{
				ResourceAttributes: &authv1.ResourceAttributes{
					Namespace: "kube-system",
					Verb:      "get",
					Resource:  "pods",
				},
				User: blueUser,
			},
		}

		sar, err := clientset.AuthorizationV1().
			SubjectAccessReviews().
			Create(context.TODO(), sar, metav1.CreateOptions{})

		is.NoErr(err)
		is.True(!sar.Status.Allowed)

	})

	t.Run("Developer can create CRDs aggregated with \"edit\" cluster role", func(t *testing.T) {

		sar := &authv1.SubjectAccessReview{
			Spec: authv1.SubjectAccessReviewSpec{
				ResourceAttributes: &authv1.ResourceAttributes{
					Namespace: blueNamespace,
					Verb:      "create",
					Group:     "redis.redis.opstreelabs.in",
					Resource:  "redis",
				},
				User: blueUser,
			},
		}

		sar, err := clientset.AuthorizationV1().
			SubjectAccessReviews().
			Create(context.TODO(), sar, metav1.CreateOptions{})

		is.NoErr(err)
		is.True(sar.Status.Allowed)

	})

}

func before(t *testing.T) {

	kubeconfig := fmt.Sprintf("%s/.kube/config", os.Getenv("HOME"))
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
}
