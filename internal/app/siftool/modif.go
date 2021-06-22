// Copyright (c) 2018-2021, Sylabs Inc. All rights reserved.
// Copyright (c) 2018, Divya Cote <divya.cote@gmail.com> All rights reserved.
// Copyright (c) 2017, SingularityWare, LLC. All rights reserved.
// Copyright (c) 2017, Yannick Cote <yhcote@gmail.com> All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package siftool

import (
	"fmt"
	"io"
	"log"

	uuid "github.com/satori/go.uuid"
	"github.com/sylabs/sif/pkg/sif"
)

// New creates a new empty SIF file.
func New(path string) error {
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}

	cinfo := sif.CreateInfo{
		Pathname:   path,
		Launchstr:  sif.HdrLaunch,
		Sifversion: sif.HdrVersion,
		ID:         id,
	}

	_, err = sif.CreateContainer(cinfo)
	return err
}

// AddOptions contains the options when adding a section to a SIF file.
type AddOptions struct {
	Datatype   sif.Datatype
	Parttype   sif.Parttype
	Partfs     sif.Fstype
	Partarch   string
	Signhash   sif.Hashtype
	Signentity string
	Groupid    uint32
	Link       uint32
	Alignment  int
	Filename   string
	Fp         io.Reader
}

// Add adds a data object to a SIF file.
func Add(path string, opts AddOptions) error {
	input := sif.DescriptorInput{
		Datatype:  opts.Datatype,
		Groupid:   sif.DescrGroupMask | opts.Groupid,
		Link:      opts.Link,
		Alignment: opts.Alignment,
		Fname:     opts.Filename,
		Fp:        opts.Fp,
	}

	if opts.Datatype == sif.DataPartition {
		if err := input.SetPartExtra(opts.Partfs, opts.Parttype, opts.Partarch); err != nil {
			return err
		}
	} else if opts.Datatype == sif.DataSignature {
		if err := input.SetSignExtra(opts.Signhash, opts.Signentity); err != nil {
			return err
		}
	}

	// load SIF image file
	fimg, err := sif.LoadContainer(path, false)
	if err != nil {
		return err
	}
	defer func() {
		if err := fimg.UnloadContainer(); err != nil {
			log.Printf("Error unloading container: %v", err)
		}
	}()

	// add new data object to SIF file
	return fimg.AddObject(input)
}

// Del deletes a specified object descriptor and data from the SIF file.
func Del(path string, id uint32) error {
	fimg, err := sif.LoadContainer(path, false)
	if err != nil {
		return err
	}
	defer func() {
		if err := fimg.UnloadContainer(); err != nil {
			log.Printf("Error unloading container: %v", err)
		}
	}()

	for _, v := range fimg.DescrArr {
		if !v.Used {
			continue
		} else if v.ID == id {
			return fimg.DeleteObject(id, 0)
		}
	}

	return fmt.Errorf("descriptor not in range or currently unused")
}

// Setprim sets the primary system partition of the SIF file.
func Setprim(path string, id uint32) error {
	fimg, err := sif.LoadContainer(path, false)
	if err != nil {
		return err
	}
	defer func() {
		if err := fimg.UnloadContainer(); err != nil {
			log.Printf("Error unloading container: %v", err)
		}
	}()

	for _, v := range fimg.DescrArr {
		if !v.Used {
			continue
		} else if v.ID == id {
			return fimg.SetPrimPart(id)
		}
	}

	return fmt.Errorf("descriptor not in range or currently unused")
}
