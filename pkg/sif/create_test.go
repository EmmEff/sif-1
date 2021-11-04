// Copyright (c) 2018-2021, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package sif

import (
	"errors"
	"os"
	"testing"

	"github.com/sebdah/goldie/v2"
)

func TestNextAligned(t *testing.T) {
	cases := []struct {
		name     string
		offset   int64
		align    int
		expected int64
	}{
		{name: "align 0 to 0", offset: 0, align: 0, expected: 0},
		{name: "align 1 to 0", offset: 1, align: 0, expected: 1},
		{name: "align 0 to 1024", offset: 0, align: 1024, expected: 0},
		{name: "align 1 to 1024", offset: 1, align: 1024, expected: 1024},
		{name: "align 1023 to 1024", offset: 1023, align: 1024, expected: 1024},
		{name: "align 1024 to 1024", offset: 1024, align: 1024, expected: 1024},
		{name: "align 1025 to 1024", offset: 1025, align: 1024, expected: 2048},
	}

	for _, tc := range cases {
		actual := nextAligned(tc.offset, tc.align)
		if actual != tc.expected {
			t.Errorf("nextAligned case: %q, offset: %d, align: %d, expecting: %d, actual: %d\n",
				tc.name, tc.offset, tc.align, tc.expected, actual)
		}
	}
}

func TestCreateContainer(t *testing.T) {
	tests := []struct {
		name    string
		opts    []CreateOpt
		wantErr error
	}{
		{
			name: "ErrInsufficientCapacity",
			opts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptorCapacity(0),
				OptCreateWithDescriptors(
					getDescriptorInput(t, DataGeneric, []byte{0xfa, 0xce}),
				),
			},
			wantErr: errInsufficientCapacity,
		},
		{
			name: "Empty",
			opts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
			},
		},
		{
			name: "LaunchScript",
			opts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithLaunchScript("#!/usr/bin/env launch-script\n"),
			},
		},
		{
			name: "EmptyCloseOnUnload",
			opts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithCloseOnUnload(true),
			},
		},
		{
			name: "EmptyDescriptorLimitedCapacity",
			opts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptorCapacity(1),
			},
		},
		{
			name: "OneDescriptor",
			opts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptors(
					getDescriptorInput(t, DataGeneric, []byte{0xfa, 0xce}),
				),
			},
		},
		{
			name: "TwoDescriptors",
			opts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptors(
					getDescriptorInput(t, DataGeneric, []byte{0xfa, 0xce}),
					getDescriptorInput(t, DataPartition, []byte{0xfe, 0xed},
						OptPartitionMetadata(FsSquash, PartPrimSys, "386"),
					),
				),
			},
		},
		{
			name: "TwoDescriptorsNotAligned",
			opts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptors(
					getDescriptorInput(t, DataGeneric, []byte{0xfa, 0xce},
						OptObjectAlignment(0),
					),
					getDescriptorInput(t, DataPartition, []byte{0xfe, 0xed},
						OptPartitionMetadata(FsSquash, PartPrimSys, "386"),
						OptObjectAlignment(0),
					),
				),
			},
		},
		{
			name: "TwoDescriptorsAligned",
			opts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptors(
					getDescriptorInput(t, DataGeneric, []byte{0xfa, 0xce},
						OptObjectAlignment(4),
					),
					getDescriptorInput(t, DataPartition, []byte{0xfe, 0xed},
						OptPartitionMetadata(FsSquash, PartPrimSys, "386"),
						OptObjectAlignment(4),
					),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b Buffer

			f, err := CreateContainer(&b, tt.opts...)

			if got, want := err, tt.wantErr; !errors.Is(got, want) {
				t.Fatalf("got error %v, want %v", got, want)
			}

			if err == nil {
				if err := f.UnloadContainer(); err != nil {
					t.Error(err)
				}

				g := goldie.New(t, goldie.WithTestNameForDir(true))
				g.Assert(t, tt.name, b.Bytes())
			}
		})
	}
}

func TestCreateContainerAtPath(t *testing.T) {
	tests := []struct {
		name    string
		opts    []CreateOpt
		wantErr error
	}{
		{
			name: "ErrInsufficientCapacity",
			opts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptorCapacity(0),
				OptCreateWithDescriptors(
					getDescriptorInput(t, DataGeneric, []byte{0xfa, 0xce}),
				),
			},
			wantErr: errInsufficientCapacity,
		},
		{
			name: "Empty",
			opts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
			},
		},
		{
			name: "EmptyDescriptorLimitedCapacity",
			opts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptorCapacity(1),
			},
		},
		{
			name: "OneDescriptor",
			opts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptors(
					getDescriptorInput(t, DataGeneric, []byte{0xfa, 0xce}),
				),
			},
		},
		{
			name: "TwoDescriptors",
			opts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptors(
					getDescriptorInput(t, DataGeneric, []byte{0xfa, 0xce}),
					getDescriptorInput(t, DataPartition, []byte{0xfe, 0xed},
						OptPartitionMetadata(FsSquash, PartPrimSys, "386"),
					),
				),
			},
		},
		{
			name: "TwoDescriptorsNotAligned",
			opts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptors(
					getDescriptorInput(t, DataGeneric, []byte{0xfa, 0xce},
						OptObjectAlignment(0),
					),
					getDescriptorInput(t, DataPartition, []byte{0xfe, 0xed},
						OptPartitionMetadata(FsSquash, PartPrimSys, "386"),
						OptObjectAlignment(0),
					),
				),
			},
		},
		{
			name: "TwoDescriptorsAligned",
			opts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptors(
					getDescriptorInput(t, DataGeneric, []byte{0xfa, 0xce},
						OptObjectAlignment(4),
					),
					getDescriptorInput(t, DataPartition, []byte{0xfe, 0xed},
						OptPartitionMetadata(FsSquash, PartPrimSys, "386"),
						OptObjectAlignment(4),
					),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf, err := os.CreateTemp("", "sif-test-*")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tf.Name())
			tf.Close()

			f, err := CreateContainerAtPath(tf.Name(), tt.opts...)

			if got, want := err, tt.wantErr; !errors.Is(got, want) {
				t.Fatalf("got error %v, want %v", got, want)
			}

			if err == nil {
				if err := f.UnloadContainer(); err != nil {
					t.Error(err)
				}

				b, err := os.ReadFile(tf.Name())
				if err != nil {
					t.Fatal(err)
				}

				g := goldie.New(t, goldie.WithTestNameForDir(true))
				g.Assert(t, tt.name, b)
			}
		})
	}
}

func TestAddObject(t *testing.T) {
	tests := []struct {
		name       string
		createOpts []CreateOpt
		di         DescriptorInput
		addOpts    []AddOpt
		wantErr    error
	}{
		{
			name: "ErrInsufficientCapacity",
			createOpts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptorCapacity(0),
			},
			di:      getDescriptorInput(t, DataGeneric, []byte{0xfe, 0xed}),
			wantErr: errInsufficientCapacity,
		},
		{
			name: "ErrPrimaryPartition",
			createOpts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptors(
					getDescriptorInput(t, DataPartition, []byte{0xfa, 0xce},
						OptPartitionMetadata(FsSquash, PartPrimSys, "386"),
					),
				),
			},
			di: getDescriptorInput(t, DataPartition, []byte{0xfe, 0xed},
				OptPartitionMetadata(FsSquash, PartPrimSys, "amd64"),
			),
			wantErr: errPrimaryPartition,
		},
		{
			name: "Empty",
			createOpts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
			},
			di: getDescriptorInput(t, DataGeneric, []byte{0xfa, 0xce}),
			addOpts: []AddOpt{
				OptAddWithTime(testTime),
			},
		},
		{
			name: "EmptyNotAligned",
			createOpts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
			},
			di: getDescriptorInput(t, DataGeneric, []byte{0xfa, 0xce},
				OptObjectAlignment(0),
			),
			addOpts: []AddOpt{
				OptAddWithTime(testTime),
			},
		},
		{
			name: "EmptyAligned",
			createOpts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
			},
			di: getDescriptorInput(t, DataGeneric, []byte{0xfa, 0xce},
				OptObjectAlignment(128),
			),
			addOpts: []AddOpt{
				OptAddWithTime(testTime),
			},
		},
		{
			name: "NotEmpty",
			createOpts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptors(
					getDescriptorInput(t, DataGeneric, []byte{0xfa, 0xce}),
				),
			},
			di: getDescriptorInput(t, DataPartition, []byte{0xfe, 0xed},
				OptPartitionMetadata(FsSquash, PartPrimSys, "386"),
			),
			addOpts: []AddOpt{
				OptAddWithTime(testTime),
			},
		},
		{
			name: "NotEmptyNotAligned",
			createOpts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptors(
					getDescriptorInput(t, DataGeneric, []byte{0xfa, 0xce}),
				),
			},
			di: getDescriptorInput(t, DataPartition, []byte{0xfe, 0xed},
				OptPartitionMetadata(FsSquash, PartPrimSys, "386"),
				OptObjectAlignment(0),
			),
			addOpts: []AddOpt{
				OptAddWithTime(testTime),
			},
		},
		{
			name: "NotEmptyAligned",
			createOpts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptors(
					getDescriptorInput(t, DataGeneric, []byte{0xfa, 0xce}),
				),
			},
			di: getDescriptorInput(t, DataPartition, []byte{0xfe, 0xed},
				OptPartitionMetadata(FsSquash, PartPrimSys, "386"),
				OptObjectAlignment(128),
			),
			addOpts: []AddOpt{
				OptAddWithTime(testTime),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b Buffer

			f, err := CreateContainer(&b, tt.createOpts...)
			if err != nil {
				t.Fatal(err)
			}

			if got, want := f.AddObject(tt.di, tt.addOpts...), tt.wantErr; !errors.Is(got, want) {
				t.Errorf("got error %v, want %v", got, want)
			}

			if err := f.UnloadContainer(); err != nil {
				t.Error(err)
			}

			g := goldie.New(t, goldie.WithTestNameForDir(true))
			g.Assert(t, tt.name, b.Bytes())
		})
	}
}

func TestDeleteObject(t *testing.T) {
	tests := []struct {
		name       string
		createOpts []CreateOpt
		id         uint32
		opts       []DeleteOpt
		wantErr    error
	}{
		{
			name: "ErrObjectNotFound",
			createOpts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
			},
			id:      1,
			wantErr: ErrObjectNotFound,
		},
		{
			name: "Zero",
			createOpts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptors(
					getDescriptorInput(t, DataGeneric, []byte{0xfa, 0xce}),
				),
			},
			id: 1,
			opts: []DeleteOpt{
				OptDeleteZero(true),
				OptDeleteWithTime(testTime),
			},
		},
		{
			name: "Compact",
			createOpts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptors(
					getDescriptorInput(t, DataGeneric, []byte{0xfa, 0xce}),
				),
			},
			id: 1,
			opts: []DeleteOpt{
				OptDeleteCompact(true),
				OptDeleteWithTime(testTime),
			},
		},
		{
			name: "ZeroCompact",
			createOpts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptors(
					getDescriptorInput(t, DataGeneric, []byte{0xfa, 0xce}),
				),
			},
			id: 1,
			opts: []DeleteOpt{
				OptDeleteZero(true),
				OptDeleteCompact(true),
				OptDeleteWithTime(testTime),
			},
		},
		{
			name: "PrimaryPartition",
			createOpts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptors(
					getDescriptorInput(t, DataPartition, []byte{0xfa, 0xce},
						OptPartitionMetadata(FsSquash, PartPrimSys, "386"),
					),
				),
			},
			id: 1,
			opts: []DeleteOpt{
				OptDeleteWithTime(testTime),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b Buffer

			f, err := CreateContainer(&b, tt.createOpts...)
			if err != nil {
				t.Fatal(err)
			}

			if got, want := f.DeleteObject(tt.id, tt.opts...), tt.wantErr; !errors.Is(got, want) {
				t.Errorf("got error %v, want %v", got, want)
			}

			if err := f.UnloadContainer(); err != nil {
				t.Error(err)
			}

			g := goldie.New(t, goldie.WithTestNameForDir(true))
			g.Assert(t, tt.name, b.Bytes())
		})
	}
}

func TestSetPrimPart(t *testing.T) {
	tests := []struct {
		name       string
		createOpts []CreateOpt
		id         uint32
		opts       []SetOpt
		wantErr    error
	}{
		{
			name: "ErrObjectNotFound",
			createOpts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
			},
			id:      1,
			wantErr: ErrObjectNotFound,
		},
		{
			name: "One",
			createOpts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptors(
					getDescriptorInput(t, DataPartition, []byte{0xfa, 0xce},
						OptPartitionMetadata(FsRaw, PartSystem, "386"),
					),
				),
			},
			id: 1,
			opts: []SetOpt{
				OptSetWithTime(testTime),
			},
		},
		{
			name: "Two",
			createOpts: []CreateOpt{
				OptCreateWithID(testID),
				OptCreateWithTime(testTime),
				OptCreateWithDescriptors(
					getDescriptorInput(t, DataPartition, []byte{0xfa, 0xce},
						OptPartitionMetadata(FsRaw, PartPrimSys, "386"),
					),
					getDescriptorInput(t, DataPartition, []byte{0xfe, 0xed},
						OptPartitionMetadata(FsRaw, PartSystem, "amd64"),
					),
				),
			},
			id: 2,
			opts: []SetOpt{
				OptSetWithTime(testTime),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b Buffer

			f, err := CreateContainer(&b, tt.createOpts...)
			if err != nil {
				t.Fatal(err)
			}

			if got, want := f.SetPrimPart(tt.id, tt.opts...), tt.wantErr; !errors.Is(got, want) {
				t.Errorf("got error %v, want %v", got, want)
			}

			if err := f.UnloadContainer(); err != nil {
				t.Error(err)
			}

			g := goldie.New(t, goldie.WithTestNameForDir(true))
			g.Assert(t, tt.name, b.Bytes())
		})
	}
}
