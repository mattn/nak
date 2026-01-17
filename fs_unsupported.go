//go:build openbsd || darwin || riscv64 || (windows && !cgo)

package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

var fsCmd = &cli.Command{
	Name:                      "fs",
	Usage:                     "mount a FUSE filesystem that exposes Nostr events as files.",
	Description:               `doesn't work on OpenBSD, Darwin, or RISC-V 64-bit.`,
	DisableSliceFlagSeparator: true,
	Action: func(ctx context.Context, c *cli.Command) error {
		return fmt.Errorf("FUSE filesystem is not supported on this platform (OpenBSD, Darwin, or RISC-V)")
	},
}
