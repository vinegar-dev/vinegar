package gpu

import (
	"errors"

	"github.com/vinegarhq/vinegar/gpu/target"
)

func HandleGpu(target target.TargetGPU, env map[string]string, isVulkan bool) (map[string]string, error) {
	if target.Id == -1 {
		return env, nil
	}

	gpus := GetSystemGPUs()

	if target.Prime {
		allowed, err := PrimeIsAllowed(gpus, isVulkan)
		if err != nil {
			return env, err
		}
		if !allowed {
			return env, nil
		}
	}

	if len(gpus) < target.Id+1 {
		return env, errors.New("gpu not found")
	}

	gpu := gpus[target.Id]

	env = gpu.ApplyEnv(env)

	return env, nil
}