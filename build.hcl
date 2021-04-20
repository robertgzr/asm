variable "TAG" {
	default = "latest"
}

group "default" {
	targets = ["lint", "binary"]
}

target "lint" {
	target = "lint"
}

target "binary" {
	target = "binary"
	output = [ "." ]
}

target "image" {
	target = "image"
	output = [
		"type=image,name=docker.io/robertgzr/asm:${TAG},push=false",
	]
}
