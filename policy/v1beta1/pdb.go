package v1beta1

import (
	"fmt"
	types2 "github.com/appscode/go/types"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	policy "k8s.io/api/policy/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	kutil "kmodules.xyz/client-go"
	"reflect"
	"time"
)

func CreateOrPatchPodDisruptionBudget(c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*policy.PodDisruptionBudget) *policy.PodDisruptionBudget) (*policy.PodDisruptionBudget, kutil.VerbType, error) {
	cur, err := c.PolicyV1beta1().PodDisruptionBudgets(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating PodDisruptionBudget %s/%s.", meta.Namespace, meta.Name)
		out, err := c.PolicyV1beta1().PodDisruptionBudgets(meta.Namespace).Create(transform(&policy.PodDisruptionBudget{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PodDisruptionBudget",
				APIVersion: policy.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}))
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchPodDisruptionBudget(c, cur, transform)
}

func PatchPodDisruptionBudget(c kubernetes.Interface, cur *policy.PodDisruptionBudget, transform func(*policy.PodDisruptionBudget) *policy.PodDisruptionBudget) (*policy.PodDisruptionBudget, kutil.VerbType, error) {
	return PatchPodDisruptionBudgetObject(c, cur, transform(cur.DeepCopy()))
}

func PatchPodDisruptionBudgetObject(c kubernetes.Interface, cur, mod *policy.PodDisruptionBudget) (*policy.PodDisruptionBudget, kutil.VerbType, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	modJson, err := json.Marshal(mod)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(curJson, modJson, policy.PodDisruptionBudget{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, kutil.VerbUnchanged, nil
	}
	glog.V(3).Infof("Patching PodDisruptionBudget %s with %s.", cur.Name, string(patch))
	out, err := c.PolicyV1beta1().PodDisruptionBudgets(cur.Namespace).Patch(cur.Name, types.StrategicMergePatchType, patch)
	return out, kutil.VerbPatched, err
}

func TryUpdatePodDisruptionBudget(c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*policy.PodDisruptionBudget) *policy.PodDisruptionBudget) (result *policy.PodDisruptionBudget, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.PolicyV1beta1().PodDisruptionBudgets(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.PolicyV1beta1().PodDisruptionBudgets(meta.Namespace).Update(transform(cur.DeepCopy()))
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update PodDisruptionBudget %s due to %v.", attempt, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = errors.Errorf("failed to update PodDisruptionBudget %s after %d attempts due to %v", meta.Name, attempt, err)
	}
	return
}

func CreateOrPatchPDB(c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*policy.PodDisruptionBudget) *policy.PodDisruptionBudget) (*policy.PodDisruptionBudget, kutil.VerbType, error) {
	cur, err := c.PolicyV1beta1().PodDisruptionBudgets(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating PDB %s/%s.", meta.Namespace, meta.Name)
		fmt.Println("===============>Create NEw")
		out, err := c.PolicyV1beta1().PodDisruptionBudgets(meta.Namespace).Create(transform(&policy.PodDisruptionBudget{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PodDisruptionBudget",
				APIVersion: policy.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}))
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	mod := transform(cur.DeepCopy())
	fmt.Println("++++++++++Current pdb = ", cur.Spec)
	fmt.Println("+++++++++++New pdb = ", mod.Spec)
	if !reflect.DeepEqual(cur.Spec , mod.Spec){
		fmt.Println("===============>PDBs ain't equal")
		// PDBs dont have the specs, Specs can't be modified once created, so we have to delete first, then recreate with correct  spec
		glog.Warningf("PDB %s/%s spec is modified, deleting first.", meta.Namespace, meta.Name)
		err = c.PolicyV1beta1().PodDisruptionBudgets(meta.Namespace).Delete(meta.Name, &metav1.DeleteOptions{GracePeriodSeconds:types2.Int64P(1),})
		if err != nil {
			fmt.Println("Ordinarily, this should produce any error, err = ", err)
			return nil, kutil.VerbUnchanged, err
		}
		fmt.Println("Sleeping")
		time.Sleep(time.Second*10)
		fmt.Println("Slept")
		glog.V(3).Infof("Creating PDB %s/%s.", mod.Namespace, mod.Name)
		fmt.Println("Creating new pdb")
		out, err := c.PolicyV1beta1().PodDisruptionBudgets(meta.Namespace).Create(transform(&policy.PodDisruptionBudget{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PodDisruptionBudget",
				APIVersion: policy.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}))
		if err != nil{
			fmt.Println("Patch error = ", err)
		}
		return out, kutil.VerbPatched, err
	} else{
		fmt.Println("+++++++++>PDBs are equal err = ", err)
	}
	fmt.Println("END err= ", err)
	return cur, "unchanged", err
}
