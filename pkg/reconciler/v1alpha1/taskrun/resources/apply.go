/*
Copyright 2018 The Knative Authors.

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

package resources

import (
	"fmt"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/pipeline/pkg/templating"
)

// ApplyParameters applies the params from a TaskRun.Input.Parameters to a TaskSpec
func ApplyParameters(spec *v1alpha1.TaskSpec, tr *v1alpha1.TaskRun, defaults ...v1alpha1.TaskParam) *v1alpha1.TaskSpec {
	// This assumes that the TaskRun inputs have been validated against what the Task requests.
	replacements := map[string]string{}
	// Set all the default replacements
	for _, p := range defaults {
		if p.Default != "" {
			replacements[fmt.Sprintf("inputs.params.%s", p.Name)] = p.Default
		}
	}
	// Set and overwrite params with the ones from the TaskRun
	for _, p := range tr.Spec.Inputs.Params {
		replacements[fmt.Sprintf("inputs.params.%s", p.Name)] = p.Value
	}

	return ApplyReplacements(spec, replacements)
}

// ApplyResources applies the templating from values in resources which are referenced in spec as subitems
// of the replacementStr.
func ApplyResources(spec *v1alpha1.TaskSpec, resolvedResources map[string]v1alpha1.PipelineResourceInterface, replacementStr string) *v1alpha1.TaskSpec {
	replacements := map[string]string{}
	for name, r := range resolvedResources {
		for k, v := range r.Replacements() {
			replacements[fmt.Sprintf("%s.resources.%s.%s", replacementStr, name, k)] = v
		}
	}
	return ApplyReplacements(spec, replacements)
}

// ApplyReplacements replaces placeholders for declared parameters with the specified replacements.
func ApplyReplacements(spec *v1alpha1.TaskSpec, replacements map[string]string) *v1alpha1.TaskSpec {
	spec = spec.DeepCopy()

	// Apply variable expansion to steps fields.
	steps := spec.Steps
	for i := range steps {
		steps[i].Name = templating.ApplyReplacements(steps[i].Name, replacements)
		steps[i].Image = templating.ApplyReplacements(steps[i].Image, replacements)
		for ia, a := range steps[i].Args {
			steps[i].Args[ia] = templating.ApplyReplacements(a, replacements)
		}
		for ie, e := range steps[i].Env {
			steps[i].Env[ie].Value = templating.ApplyReplacements(e.Value, replacements)
			if steps[i].Env[ie].ValueFrom != nil {
				if e.ValueFrom.SecretKeyRef != nil {
					steps[i].Env[ie].ValueFrom.SecretKeyRef.LocalObjectReference.Name = templating.ApplyReplacements(e.ValueFrom.SecretKeyRef.LocalObjectReference.Name, replacements)
					steps[i].Env[ie].ValueFrom.SecretKeyRef.Key = templating.ApplyReplacements(e.ValueFrom.SecretKeyRef.Key, replacements)
				}
				if e.ValueFrom.ConfigMapKeyRef != nil {
					steps[i].Env[ie].ValueFrom.ConfigMapKeyRef.LocalObjectReference.Name = templating.ApplyReplacements(e.ValueFrom.ConfigMapKeyRef.LocalObjectReference.Name, replacements)
					steps[i].Env[ie].ValueFrom.ConfigMapKeyRef.Key = templating.ApplyReplacements(e.ValueFrom.ConfigMapKeyRef.Key, replacements)
				}
			}
		}
		for ie, e := range steps[i].EnvFrom {
			steps[i].EnvFrom[ie].Prefix = templating.ApplyReplacements(e.Prefix, replacements)
			if e.ConfigMapRef != nil {
				steps[i].EnvFrom[ie].ConfigMapRef.LocalObjectReference.Name = templating.ApplyReplacements(e.ConfigMapRef.LocalObjectReference.Name, replacements)
			}
			if e.SecretRef != nil {
				steps[i].EnvFrom[ie].SecretRef.LocalObjectReference.Name = templating.ApplyReplacements(e.SecretRef.LocalObjectReference.Name, replacements)
			}
		}
		steps[i].WorkingDir = templating.ApplyReplacements(steps[i].WorkingDir, replacements)
		for ic, c := range steps[i].Command {
			steps[i].Command[ic] = templating.ApplyReplacements(c, replacements)
		}
		for iv, v := range steps[i].VolumeMounts {
			steps[i].VolumeMounts[iv].Name = templating.ApplyReplacements(v.Name, replacements)
			steps[i].VolumeMounts[iv].MountPath = templating.ApplyReplacements(v.MountPath, replacements)
			steps[i].VolumeMounts[iv].SubPath = templating.ApplyReplacements(v.SubPath, replacements)
		}
	}

	// Apply variable expansion to the build's volumes
	for i, v := range spec.Volumes {
		spec.Volumes[i].Name = templating.ApplyReplacements(v.Name, replacements)
		if v.VolumeSource.ConfigMap != nil {
			spec.Volumes[i].ConfigMap.Name = templating.ApplyReplacements(v.ConfigMap.Name, replacements)
		}
		if v.VolumeSource.Secret != nil {
			spec.Volumes[i].Secret.SecretName = templating.ApplyReplacements(v.Secret.SecretName, replacements)
		}
		if v.PersistentVolumeClaim != nil {
			spec.Volumes[i].PersistentVolumeClaim.ClaimName = templating.ApplyReplacements(v.PersistentVolumeClaim.ClaimName, replacements)
		}
	}

	return spec
}
