package main

import (
	"context"
	"encoding/json"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"path/filepath"
)

var gvk = schema.GroupVersionKind{
	Group:   "config.kio.kasten.io",
	Version: "v1alpha1",
	Kind:    "Profile",
}

var gvr = schema.GroupVersionResource{
	Group:    "config.kio.kasten.io",
	Version:  "v1alpha1",
	Resource: "profiles",
}

type ProfileSpec struct {
	LocationSpec map[string]interface{} `json:"locationSpec"`
	Type         string                 `json:"type"`
}

type Profile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ProfileSpec `json:"spec,omitempty"`
}

type ProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Profile `json:"items"`
}

func deleteProfile(client dynamic.Interface, namespace string, name string) error {
	return client.Resource(gvr).Namespace(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

func patchProfile(client dynamic.Interface, namespace, name string, pt types.PatchType, data []byte) error {
	_, err := client.Resource(gvr).Namespace(namespace).Patch(context.Background(), name, pt, data, metav1.PatchOptions{})
	return err
}

func updateProfileWithYaml(client dynamic.Interface, namespace string, yamlData string) (*Profile, error) {
	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	if _, _, err := decoder.Decode([]byte(yamlData), &gvk, obj); err != nil {
		return nil, err
	}

	utd, err := client.Resource(gvr).Namespace(namespace).Get(context.Background(), obj.GetName(), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	obj.SetResourceVersion(utd.GetResourceVersion())
	utd, err = client.Resource(gvr).Namespace(namespace).Update(context.Background(), obj, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	data, err := utd.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var ct Profile
	if err := json.Unmarshal(data, &ct); err != nil {
		return nil, err
	}
	return &ct, nil
}

func createProfileWithYaml(client dynamic.Interface, namespace string, yamlData string) (*Profile, error) {
	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	if _, _, err := decoder.Decode([]byte(yamlData), &gvk, obj); err != nil {
		return nil, err
	}

	utd, err := client.Resource(gvr).Namespace(namespace).Create(context.Background(), obj, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	data, err := utd.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var ct Profile
	if err := json.Unmarshal(data, &ct); err != nil {
		return nil, err
	}
	return &ct, nil
}

func listProfiles(client dynamic.Interface, namespace string) (*ProfileList, error) {
	list, err := client.Resource(gvr).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
	println(client.Resource(gvr))
	if err != nil {
		return nil, err
	}
	data, err := list.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var ctList ProfileList
	//println(ctList)
	if err := json.Unmarshal(data, &ctList); err != nil {
		return nil, err
	}
	return &ctList, nil
}

func getProfile(client dynamic.Interface, namespace string, name string) (*Profile, error) {
	utd, err := client.Resource(gvr).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	data, err := utd.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var ct Profile
	if err := json.Unmarshal(data, &ct); err != nil {
		return nil, err
	}
	return &ct, nil
}

func main() {
	//1. Set Kubeconfig
	kubeconfig := filepath.Join("c:", "\\goproject", "config")
	fmt.Println(kubeconfig)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	crudOperations := ""
	fmt.Println("please input CRUD OPERATION, GET LIST CREATE PATCH DELETE UPDATE")
	fmt.Scanln(&crudOperations)

	switch crudOperations {
	case "GET":
		//get specific profile by name filter from CRD profiles.config.kio.kasten.io
		ct, err := getProfile(client, "kasten-io", "cos1")
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s %s %s %s\n", ct.Namespace, ct.Name, ct.Spec.LocationSpec, ct.Spec.Type)
	case "LIST":
		// get profiles list from CRD profiles.config.kio.kasten.io
		list, err := listProfiles(client, "kasten-io")
		if err != nil {
			panic(err)
		}
		for _, t := range list.Items {
			fmt.Printf("%s %s %s %s\n", t.Namespace, t.Name, t.Spec.LocationSpec, t.Spec.Type)
		}
	case "DELETE":
		if err := deleteProfile(client, "kasten-io", "cos2"); err != nil {
			panic(err)
		}
	case "PATCH":
		// patch profile from CR profiles.config.kio.kasten.io
		patchData := []byte(`{"spec": {"locationSpec": {"credential": {"secret": {"name": "k10secret-mars"}}}}}`)
		//patchData := []byte(`{"metadata": {"name":"cos1-patch"}}`)
		if err := patchProfile(client, "kasten-io", "cos1", types.MergePatchType, patchData); err != nil {
			panic(err)
		}
	case "UPDATE":
		//update profile
		updateData := `apiVersion: config.kio.kasten.io/v1alpha1
kind: Profile
metadata:
  name: cos1
  namespace: kasten-io
spec:
  type: Location
  locationSpec:
    credential:
      secretType: AwsAccessKey
      secret:
        apiVersion: v1
        kind: Secret
        name: k10secret-mars
        namespace: kasten-io
    type: ObjectStore
    objectStore:
      endpoint: https://cos.ap-chengdu.myqcloud.com
      name: kasten-XXXXX
      objectStoreType: S3
      region: ap-chengdu`
		ct, err := updateProfileWithYaml(client, "kasten-io", updateData)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s %s %s %s\n", ct.Namespace, ct.Name, ct.Spec.LocationSpec, ct.Spec.Type)

	case "CREATE":
		// Create profiles with Yaml
		createData := `apiVersion: config.kio.kasten.io/v1alpha1
kind: Profile
metadata:
  name: cos1
  namespace: kasten-io
spec:
  type: Location
  locationSpec:
    credential:
      secretType: AwsAccessKey
      secret:
        apiVersion: v1
        kind: Secret
        name: k10secret-wshlm
        namespace: kasten-io
    type: ObjectStore
    objectStore:
      endpoint: https://cos.ap-chengdu.myqcloud.com
      name: kasten-XXXXXX
      objectStoreType: S3
      region: ap-chengdu`
		fmt.Printf(createData)
		ct, err := createProfileWithYaml(client, "kasten-io", createData)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s %s %s %s\n", ct.Namespace, ct.Name, ct.Spec.LocationSpec, ct.Spec.Type)
	}

}
