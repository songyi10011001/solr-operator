/*
Copyright 2019 Bloomberg Finance LP.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"testing"

	solr "github.com/bloomberg/solr-operator/api/v1beta1"
	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ reconcile.Reconciler = &SolrCloudReconciler{}

var (
	expectedCloudRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: "foo-clo", Namespace: "default"}}
	cloudSsKey           = types.NamespacedName{Name: "foo-clo-solrcloud", Namespace: "default"}
	cloudCsKey           = types.NamespacedName{Name: "foo-clo-solrcloud-common", Namespace: "default"}
	cloudHsKey           = types.NamespacedName{Name: "foo-clo-solrcloud-headless", Namespace: "default"}
	cloudIKey            = types.NamespacedName{Name: "foo-clo-solrcloud-common", Namespace: "default"}
)

func TestCloudReconcile(t *testing.T) {
	SetIngressBaseUrl("")
	UseEtcdCRD(false)
	UseZkCRD(true)
	g := gomega.NewGomegaWithT(t)
	instance := &solr.SolrCloud{
		ObjectMeta: metav1.ObjectMeta{Name: expectedCloudRequest.Name, Namespace: expectedCloudRequest.Namespace},
		Spec: solr.SolrCloudSpec{
			ZookeeperRef: &solr.ZookeeperRef{
				ConnectionInfo: &solr.ZookeeperConnectionInfo{
					InternalConnectionString: "host:7271",
				},
			},
		},
	}

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(testCfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	testClient = mgr.GetClient()

	solrCloudReconciler := &SolrCloudReconciler{
		Client: testClient,
		Log:    ctrl.Log.WithName("controllers").WithName("SolrCloud"),
	}
	newRec, requests := SetupTestReconcile(solrCloudReconciler)
	g.Expect(solrCloudReconciler.SetupWithManagerAndReconciler(mgr, newRec)).NotTo(gomega.HaveOccurred())

	stopMgr, mgrStopped := StartTestManager(mgr, g)

	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	// Create the SolrCloud object and expect the Reconcile and StatefulSet to be created
	err = testClient.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer testClient.Delete(context.TODO(), instance)
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedCloudRequest)))

	// Check the statefulSet
	statefulSet := expectStatefulSet(t, g, requests, expectedCloudRequest, cloudSsKey)

	// Check the client Service
	expectService(t, g, requests, expectedCloudRequest, cloudCsKey, statefulSet.Spec.Template.Labels)

	// Check the headless Service
	expectService(t, g, requests, expectedCloudRequest, cloudHsKey, statefulSet.Spec.Template.Labels)

	// Check the ingress
	expectNoIngress(g, requests, cloudIKey)
}

func TestCloudReconcileWithIngress(t *testing.T) {
	SetIngressBaseUrl("ing.base.domain")
	UseEtcdCRD(false)
	UseZkCRD(true)
	g := gomega.NewGomegaWithT(t)
	instance := &solr.SolrCloud{
		ObjectMeta: metav1.ObjectMeta{Name: expectedCloudRequest.Name, Namespace: expectedCloudRequest.Namespace},
		Spec: solr.SolrCloudSpec{
			ZookeeperRef: &solr.ZookeeperRef{
				ConnectionInfo: &solr.ZookeeperConnectionInfo{
					InternalConnectionString: "host:7271",
				},
			},
		},
	}

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(testCfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	testClient = mgr.GetClient()

	solrCloudReconciler := &SolrCloudReconciler{
		Client: testClient,
		Log:    ctrl.Log.WithName("controllers").WithName("SolrCloud"),
	}
	newRec, requests := SetupTestReconcile(solrCloudReconciler)
	g.Expect(solrCloudReconciler.SetupWithManagerAndReconciler(mgr, newRec)).NotTo(gomega.HaveOccurred())

	stopMgr, mgrStopped := StartTestManager(mgr, g)

	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	// Create the SolrCloud object and expect the Reconcile and StatefulSet to be created
	err = testClient.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer testClient.Delete(context.TODO(), instance)
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedCloudRequest)))

	// Check the statefulSet
	statefulSet := expectStatefulSet(t, g, requests, expectedCloudRequest, cloudSsKey)

	// Check the client Service
	expectService(t, g, requests, expectedCloudRequest, cloudCsKey, statefulSet.Spec.Template.Labels)

	// Check the headless Service
	expectService(t, g, requests, expectedCloudRequest, cloudHsKey, statefulSet.Spec.Template.Labels)

	// Check the ingress
	expectIngress(g, requests, expectedCloudRequest, cloudIKey)
}