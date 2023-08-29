package main

import (
	"fmt"
	"github.com/ImSingee/kitty/internal/lib/tl"
	"github.com/spf13/cobra"
	"time"
)

func init() {
	enableSubSub4 := false

	extensions = append(extensions,
		&cobra.Command{
			Use:    "tl",
			Hidden: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				return tl.New([]*tl.Task{
					{
						Title: "task1",
						Run: func(callback tl.TaskCallback) error {
							time.Sleep(1 * time.Second)
							return nil
						},
					},
					{
						Title: "task2",
						Run: func(callback tl.TaskCallback) error {
							callback.Skip("for test")
							return nil
						},
					},
					{
						Title: "task3",
						Run: func(callback tl.TaskCallback) error {
							callback.AddSubTaskList(tl.NewTaskList(nil, tl.WithExitOnError(false)))
							callback.AddSubTask(&tl.Task{
								Title: "subtask1",
								Run: func(callback tl.TaskCallback) error {
									time.Sleep(1 * time.Second)
									return nil
								},
							})
							callback.AddSubTask(&tl.Task{
								Title: "subtask2",
								Run: func(callback tl.TaskCallback) error {
									return fmt.Errorf("test")
								},
							})
							callback.AddSubTask(&tl.Task{
								Title: "subtask3",
								Run: func(callback tl.TaskCallback) error {
									callback.AddSubTaskList(tl.NewTaskList(
										[]*tl.Task{
											{
												Title: "subsub1 (disabled)",
												Run: func(callback tl.TaskCallback) error {
													time.Sleep(1 * time.Minute)
													return nil
												},
												Enable: func() bool {
													return false
												},
											},
											{
												Title: "subsub2",
												Run: func(callback tl.TaskCallback) error {
													time.Sleep(1 * time.Second)
													return nil
												},
											},
											{
												Title: "subsub3",
												Run: func(callback tl.TaskCallback) error {
													time.Sleep(1 * time.Second)
													return nil
												},
											},
											{
												Title: "subsub4 (disabled first, then enable)",
												Run: func(callback tl.TaskCallback) error {
													time.Sleep(1 * time.Second)
													return nil
												},
												Enable: func() bool {
													if !enableSubSub4 {
														enableSubSub4 = true
														return false
													}

													return true
												},
											},
										},
									))
									return nil
								},
							})
							return nil
						},
						Options: []tl.OptionApplier{
							tl.WithExitOnError(false),
						},
					},
					{
						Title: "task4",
						Run: func(callback tl.TaskCallback) error {
							return fmt.Errorf("test")
						},
						Options: []tl.OptionApplier{
							tl.WithExitOnError(false),
						},
					},
					{
						Title: "task5",
						Run: func(callback tl.TaskCallback) error {
							time.Sleep(1 * time.Second)
							return nil
						},
					},
				},
				//tl.WithExitOnError(false),
				).Run()
			},
		},
	)
}
