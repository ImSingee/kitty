package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"

	"github.com/ImSingee/kitty/internal/lib/tl"
)

func init() {
	enableSubSub4 := false
	subsub6Error := false

	_ = &cobra.Command{
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
											Title: "subsub2 (wait until ctrl+c)",
											Run: func(callback tl.TaskCallback) error {
												signalCh := make(chan os.Signal, 1)
												signal.Notify(signalCh, os.Interrupt)

												<-signalCh
												return nil
											},
											Enable: func() bool {
												return false
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
										{
											Title: "subsub5 (hide after run)",
											Run: func(callback tl.TaskCallback) error {
												callback.Hide()
												return nil
											},
										},
										{
											Title: "subsub6 (panic)",
											Run: func(callback tl.TaskCallback) error {
												panic("test")
												return nil
											},
											PostRun: func(result *tl.Result) {
												if result.Error {
													subsub6Error = true
												}
											},
										},
										{
											Title: "subsub7 (only run if subsub6 is error)",
											Run: func(callback tl.TaskCallback) error {
												return nil
											},
											Enable: func() bool {
												return subsub6Error
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
						callback.AddSubTaskList(tl.NewTaskList(nil))
						return nil
					},
				},
				{
					Title: "task6",
					Run: func(callback tl.TaskCallback) error {
						time.Sleep(1 * time.Second)
						return nil
					},
				},
			},
			//tl.WithExitOnError(false),
			).Run()
		},
	}
}
