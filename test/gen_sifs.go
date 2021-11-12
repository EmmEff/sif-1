// Copyright (c) 2020-2021, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the LICENSE.md file
// distributed with the sources of this project regarding your rights to use or distribute this
// software.

package main

import (
	"bytes"
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/sylabs/sif/v2/pkg/integrity"
	"github.com/sylabs/sif/v2/pkg/sif"
)

var errUnexpectedNumEntities = errors.New("unexpected number of entities")

func getEntity() (*openpgp.Entity, error) {
	f, err := os.Open(filepath.Join("keys", "private.asc"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	el, err := openpgp.ReadArmoredKeyRing(f)
	if err != nil {
		return nil, err
	}

	if len(el) != 1 {
		return nil, errUnexpectedNumEntities
	}
	return el[0], nil
}

func generateImages() error {
	e, err := getEntity()
	if err != nil {
		return err
	}

	objectGenericJSON := func() (sif.DescriptorInput, error) {
		return sif.NewDescriptorInput(sif.DataGenericJSON,
			bytes.NewReader([]byte{0x7b, 0x7d}),
			sif.OptObjectName("data.json"),
		)
	}

	objectCryptoMessage := func() (sif.DescriptorInput, error) {
		return sif.NewDescriptorInput(sif.DataCryptoMessage,
			bytes.NewReader([]byte{0xfe, 0xfe, 0xf0, 0xf0}),
			sif.OptCryptoMessageMetadata(sif.FormatOpenPGP, sif.MessageClearSignature),
		)
	}

	partSystem := func() (sif.DescriptorInput, error) {
		return sif.NewDescriptorInput(sif.DataPartition,
			bytes.NewReader([]byte{0xfa, 0xce, 0xfe, 0xed}),
			sif.OptPartitionMetadata(sif.FsRaw, sif.PartSystem, "386"),
		)
	}

	partPrimSys := func() (sif.DescriptorInput, error) {
		return sif.NewDescriptorInput(sif.DataPartition,
			bytes.NewReader([]byte{0xde, 0xad, 0xbe, 0xef}),
			sif.OptPartitionMetadata(sif.FsSquash, sif.PartPrimSys, "386"),
		)
	}

	partSystemGroup2 := func() (sif.DescriptorInput, error) {
		return sif.NewDescriptorInput(sif.DataPartition,
			bytes.NewReader([]byte{0xba, 0xdd, 0xca, 0xfe}),
			sif.OptPartitionMetadata(sif.FsExt3, sif.PartSystem, "amd64"),
			sif.OptGroupID(2),
		)
	}

	images := []struct {
		path  string
		diFns []func() (sif.DescriptorInput, error)
		opts  []sif.CreateOpt
		sign  bool
	}{
		// Images with no objects.
		{
			path: "empty.sif",
		},
		{
			path: "empty-id.sif",
			opts: []sif.CreateOpt{
				sif.OptCreateWithID("3fa802cc-358b-45e3-bcc0-69dc7a45f9f8"),
			},
		},
		{
			path: "empty-launch-script.sif",
			opts: []sif.CreateOpt{
				sif.OptCreateWithLaunchScript("#!/usr/bin/env run-script\n"),
			},
		},

		// Images with one data object in one group.
		{
			path: "one-object-time.sif",
			opts: []sif.CreateOpt{
				sif.OptCreateWithTime(time.Date(2020, 6, 30, 0, 1, 56, 0, time.UTC)),
			},
			diFns: []func() (sif.DescriptorInput, error){
				objectGenericJSON,
			},
		},
		{
			path: "one-object-generic-json.sif",
			diFns: []func() (sif.DescriptorInput, error){
				objectGenericJSON,
			},
		},
		{
			path: "one-object-crypt-message.sif",
			diFns: []func() (sif.DescriptorInput, error){
				objectCryptoMessage,
			},
		},

		// Images with two partitions in one group.
		{
			path: "one-group.sif",
			diFns: []func() (sif.DescriptorInput, error){
				partSystem,
				partPrimSys,
			},
		},
		{
			path: "one-group-signed.sif",
			diFns: []func() (sif.DescriptorInput, error){
				partSystem,
				partPrimSys,
			},
			sign: true,
		},

		// Images with three partitions in two groups.
		{
			path: "two-groups.sif",
			diFns: []func() (sif.DescriptorInput, error){
				partSystem,
				partPrimSys,
				partSystemGroup2,
			},
		},
		{
			path: "two-groups-signed.sif",
			diFns: []func() (sif.DescriptorInput, error){
				partSystem,
				partPrimSys,
				partSystemGroup2,
			},
			sign: true,
		},
	}

	for _, image := range images {
		dis := make([]sif.DescriptorInput, 0, len(image.diFns))
		for _, fn := range image.diFns {
			di, err := fn()
			if err != nil {
				return err
			}
			dis = append(dis, di)
		}

		opts := []sif.CreateOpt{
			sif.OptCreateDeterministic(),
			sif.OptCreateWithDescriptors(dis...),
		}
		opts = append(opts, image.opts...)

		f, err := sif.CreateContainerAtPath(filepath.Join("images", image.path), opts...)
		if err != nil {
			return err
		}
		defer func() {
			if err := f.UnloadContainer(); err != nil {
				log.Printf("failed to unload container: %v", err)
			}
		}()

		if image.sign {
			s, err := integrity.NewSigner(f,
				integrity.OptSignWithEntity(e),
				integrity.OptSignWithTime(func() time.Time { return time.Date(2020, 6, 30, 0, 1, 56, 0, time.UTC) }),
				integrity.OptSignDeterministic(),
			)
			if err != nil {
				return err
			}

			if err := s.Sign(); err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	if err := generateImages(); err != nil {
		log.Fatal(err)
	}
}
