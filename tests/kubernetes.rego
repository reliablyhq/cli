package kubernetes

violations[msg] {
  input.kind = "Namespace"
  input.metadata.name == "reliably"
  msg = "Reliably namespace is forbidden"
}

violations[msg] {
  input.kind = "Pod"
  container = input.spec.containers[_]
  endswith(container.image, ":latest")
  msg := sprintf("Latest tags are forbidden: image %v | container %v | pod %v", [container.image, container.name, input.metadata.name])
}

violations[msg] {
  input.kind = "Pod"
  container = input.spec.containers[_]
  not contains(container.image, ":")
  msg := sprintf("Untagged images are forbidden: image %v | container %v | pod %v", [container.image, container.name, input.metadata.name])
}

violations[msg] {
  input.kind == "Pod"
  container = input.spec.containers[_]
  not contains(container.image, "/")
  msg := sprintf("Image %v comes from untrusted registry | container %v | pod %v", [container.image, container.name, input.metadata.name])
}