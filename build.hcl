variable "TAG" {
	default = "latest"
}

group "default" {
	targets = ["lint", "build"]
}

target "lint" {
	target = "lint"
	output = [ "type=tar,dest=/dev/null" ]
}

target "build" {
	target = "final"
	output = [ "." ]
}

target "image" {
	target = "run"
	output = [
		"type=image,name=docker.io/robertgzr/asm:${TAG},push=false",
	]
}
