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
