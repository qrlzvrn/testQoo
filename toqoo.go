package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/urfave/cli/v2"
)

type Task struct {
	ID         string
	Content    string
	Complete   bool
	Category   string
	Deadline   string
	Importance int
}

func main() {
	app := &cli.App{
		Name:  "toqoo",
		Usage: "blablabla",
		Commands: []*cli.Command{
			{
				Name:  "add",
				Usage: "add a new task",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "category",
						Aliases: []string{"c"},
						Value:   "Purgatorium",
						Usage:   "Задает категорию для добавляемой задачи",
					},
					&cli.StringFlag{
						Name:    "deadline",
						Aliases: []string{"d"},
						Value:   "00/00",
						Usage:   "Задает дедлайн для добавляемой задачи",
					},
					&cli.IntFlag{
						Name:    "importance",
						Aliases: []string{"i"},
						Value:   0,
						Usage:   "Задает важность добавляемой задачи",
					},
					&cli.StringFlag{
						Name:    "content",
						Aliases: []string{"C"},
						Usage:   "Задает текст задачи",
					},
				},
				Action: func(c *cli.Context) error {
					//TODO Реализовать генирацию ID по аналогии с Git
					t := Task{
						ID:         "WS42K", //ID заглушка, для проверки работоспособности
						Content:    c.String("content"),
						Complete:   false,
						Category:   c.String("category"),
						Deadline:   c.String("deadline"),
						Importance: c.Int("importance"),
					}

					if t.Content == "" {
						err := fmt.Errorf("content can't be empty")
						return err
					}

					//TODO дописать регулярку для проверки правильности введеного дедлайна

					if err := addTask(t, "task.json"); err != nil {
						return err
					}

					fmt.Println("added task: ", t.Content)
					return nil

				},
			},
			{
				Name:  "complete",
				Usage: "complete task",
				Action: func(c *cli.Context) error {
					ID := c.Args().First()
					if ID == "" {
						err := fmt.Errorf("task ID cannot be empty")
						return err
					}
					if err := changeTask(ID, "Complete", "true"); err != nil {
						return err
					}
					fmt.Printf("task with id %s is complete\n", ID)
					return nil
				},
			},
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "list the tasks",
				Action: func(c *cli.Context) error {
					if c.NArg() > 0 {
						category := c.Args().First()
						if err := showTaskList(category); err != nil {
							return err
						}
					} else {
						if err := showTaskList("all"); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Name:  "reImp",
				Usage: "Позволяет переназначить Важность задачи",
				Action: func(c *cli.Context) error {
					ID := c.Args().First()
					Importance := c.Args().Get(1)
					if err := changeTask(ID, "Importance", Importance); err != nil {
						return err
					}
					fmt.Printf("importnace of task with ID %s changed on %s\n", ID, Importance)
					return nil
				},
			},
			{
				Name:  "reDead",
				Usage: "Позволяет переназначить Дедлайн задачи",
				Action: func(c *cli.Context) error {
					ID := c.Args().First()
					Deadline := c.Args().Get(1)
					if err := changeTask(ID, "Deadline", Deadline); err != nil {
						//fmt.Println("failed to change the deadline:", err)
						return err
					}
					fmt.Printf("deadline of task with ID %s changed on %s\n", ID, Deadline)

					return nil
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func openTaskFile() (*os.File, error) {
	if _, err := os.Stat("task.json"); os.IsNotExist(err) {
		return nil, err
	}
	file, err := os.Open("task.json")
	if err != nil {
		return nil, err
	}
	return file, nil
}

func addTask(t Task, filename string) error {
	j, err := json.Marshal(t)
	if err != nil {
		return err
	}

	j = append(j, "\n"...)
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	if _, err := f.Write(j); err != nil {
		return err
	}
	return nil
}

func changeTask(ID string, field string, value string) error {
	file, err := openTaskFile()
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		j := scanner.Text()
		t := Task{}
		err := json.Unmarshal([]byte(j), &t)
		if err != nil {
			return err
		}

		switch {
		case field == "Complete":
			if t.ID == ID {
				if !t.Complete {
					val, err := strconv.ParseBool(value)
					if err != nil {
						return err
					}
					t.Complete = val
				} else {
					err := fmt.Errorf("this task is already complete")
					return err
				}
			}
		case field == "Deadline":
			if t.ID == ID {
				t.Deadline = value
			}
		case field == "Importance":
			val, _ := strconv.Atoi(value)
			if t.ID == ID {
				t.Importance = val
			}
		default:
			err := fmt.Errorf("field does not exist")
			return err
		}

		addTask(t, ".tempTaskFile")
	}

	os.Rename(".tempTaskFile", "task.json")
	os.Remove(".tempTaskFile")
	return nil
}

func showTaskList(ctgry string) error {
	file, err := openTaskFile()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		j := scanner.Text()
		t := Task{}
		err := json.Unmarshal([]byte(j), &t)
		if err != nil {
			return err
		}

		if ctgry == "all" {
			if !t.Complete {
				fmt.Printf("[%s]\t -%s-\t***%s***\t\t <%s>\t !!%d\n", t.ID, t.Deadline, t.Content, t.Category, t.Importance)
			}
		} else if t.Category == ctgry {
			if !t.Complete {
				fmt.Printf("[%s]\t ***%s***\t ##%s\t !!%d\n", t.ID, t.Content, t.Deadline, t.Importance)
			}
		}

	}
	return nil
}
