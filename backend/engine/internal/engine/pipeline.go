package engine

func ExecuteSequential(steps []string) int {
  return len(steps)
}

func ExecuteParallel(steps []string) int {
  return len(steps)
}

func EvaluateCondition(flag bool) string {
  if flag {
    return "then"
  }
  return "else"
}
