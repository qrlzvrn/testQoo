package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
)

type task struct {
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
		Usage: "it is a simple cli program like todo.txt that uses json files for storing tasks.\n\nThis program can add, complete, show, and also reassign the importance and deadlines of tasks.",
		Commands: []*cli.Command{
			{
				Name:  "add",
				Usage: "add a new task",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "content",
						Aliases:  []string{"C"},
						Usage:    "set task content",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "category",
						Aliases: []string{"c"},
						Value:   "Purgatorium",
						Usage:   "set task category",
					},
					&cli.StringFlag{
						Name:    "deadline",
						Aliases: []string{"d"},
						Value:   "00/00",
						Usage:   "set task deadline",
					},
					&cli.IntFlag{
						Name:    "importance",
						Aliases: []string{"i"},
						Value:   0,
						Usage:   "set task importance",
					},
				},
				Action: func(c *cli.Context) error {
					t := task{
						ID:         makeID(c.String("content"), c.String("deadline")),
						Content:    c.String("content"),
						Complete:   false,
						Category:   c.String("category"),
						Deadline:   c.String("deadline"),
						Importance: c.Int("importance"),
					}

					if t.Content == "" {
						err := fmt.Errorf("task content can't be empty")
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
				Usage: "complete the task",
				Action: func(c *cli.Context) error {
					ID := c.Args().First()
					if ID == "" {
						err := fmt.Errorf("task ID cannot be empty")
						return err
					}
					if err := changeTask(ID, "Complete", "true"); err != nil {
						return err
					}
					fmt.Printf("task with an id %s is complete\n", ID)
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
				Usage: "reassign task importance",
				Action: func(c *cli.Context) error {
					ID := c.Args().First()
					Importance := c.Args().Get(1)
					if err := changeTask(ID, "Importance", Importance); err != nil {
						return err
					}
					fmt.Printf("importnace of task with an ID %s changed on %s\n", ID, Importance)
					return nil
				},
			},
			{
				Name:  "reDead",
				Usage: "reassign task deadline",
				Action: func(c *cli.Context) error {
					ID := c.Args().First()
					Deadline := c.Args().Get(1)
					if err := changeTask(ID, "Deadline", Deadline); err != nil {
						return err
					}
					fmt.Printf("deadline of task with an ID %s changed on %s\n", ID, Deadline)

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

func addTask(t task, filename string) error {
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

	taskExist := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		j := scanner.Text()
		t := task{}
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
					taskExist = true
				} else {
					err := fmt.Errorf("this task is already complete")
					return err
				}
			}
		case field == "Deadline":
			if t.ID == ID {
				t.Deadline = value
				taskExist = true
			}
		case field == "Importance":
			val, _ := strconv.Atoi(value)
			if t.ID == ID {
				t.Importance = val
				taskExist = true
			}
		default:
			err := fmt.Errorf("field doesn't exist")
			return err
		}
		addTask(t, ".tempTaskFile")
	}

	if taskExist == true {
		os.Rename(".tempTaskFile", "task.json")
		return nil
	} else {
		err := fmt.Errorf("task with this ID does not exist")
		os.Remove(".tempTaskFile")
		return err
	}
}

func showTaskList(ctgry string) error {
	file, err := openTaskFile()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		j := scanner.Text()
		t := task{}
		err := json.Unmarshal([]byte(j), &t)
		if err != nil {
			return err
		}

		if ctgry == "all" {
			if !t.Complete {
				fmt.Printf("[%s]\t -%s-\t %s\t\t Imp:%d\t\t  <%s>\n", t.ID, t.Deadline, t.Content, t.Importance, t.Category)
			}
		} else if t.Category == ctgry {
			if !t.Complete {
				fmt.Printf("[%s]\t -%s-\t %s \t\t Imp: %d\n", t.ID, t.Deadline, t.Content, t.Importance)
			}
		}

	}
	return nil
}

func makeID(content, deadline string) string {

	str := content + deadline
	sha256 := sha256.Sum256([]byte(str))
	shaID := sha256[:4]

	//Json не воспринимает бинарные данные, поэтому переводим получившийся срез байт в строку
	var b strings.Builder

	for _, i := range shaID {
		fmt.Fprintf(&b, "%x", i)
	}

	return b.String()
}
