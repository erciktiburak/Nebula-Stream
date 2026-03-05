package runtime

func RuntimeName() string {
  return "wasmtime"
}

func SandboxEnabled() bool {
  return true
}

type StepData map[string][]byte
