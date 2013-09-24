package main

type TextScanner interface {
	Scan() bool
	Text() string
}
