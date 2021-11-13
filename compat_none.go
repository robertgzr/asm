//go:build !balena
// +build !balena

package asm

import (
	"github.com/docker/buildx/bake"
	"github.com/docker/buildx/build"
)

func parseBalena(m targetMap) error {
	return nil
}

func resolveBalena(bo map[string]build.Options, m map[string]*bake.Target) error {
	return nil
}
