#!/bin/bash

# see https://github.com/openshift/origin/blob/master/hack/text.sh

# This file contains helpful aliases for manipulating the output text to the terminal as
# well as functions for one-command augmented printing.

# oscan::text::reset resets the terminal output to default if it is called in a TTY
function oscan::text::reset() {
	if [ -t 1 ]; then
		tput sgr0
	fi
}

# oscan::text::bold sets the terminal output to bold text if it is called in a TTY
function oscan::text::bold() {
	if [ -t 1 ]; then
		tput bold
	fi
}

# oscan::text::red sets the terminal output to red text if it is called in a TTY
function oscan::text::red() {
	if [ -t 1 ]; then
		tput setaf 1
	fi
}

# oscan::text::green sets the terminal output to green text if it is called in a TTY
function oscan::text::green() {
	if [ -t 1 ]; then
		tput setaf 2
	fi
}

# oscan::text::blue sets the terminal output to blue text if it is called in a TTY
function oscan::text::blue() {
	if [ -t 1 ]; then
		tput setaf 4
	fi
}

# oscan::text::yellow sets the terminal output to yellow text if it is called in a TTY
function oscan::text::yellow() {
	if [ -t 1 ]; then
		tput setaf 11
	fi
}

# oscan::text::clear_last_line clears the text from the last line of output to the
# terminal and leaves the cursor on that line to allow for overwriting that text
# if it is called in a TTY
function oscan::text::clear_last_line() {
	if [ -t 1 ]; then 
		tput cuu 1
		tput el
	fi
}

# oscan::text::print_bold prints all input in bold text
function oscan::text::print_bold() {
	oscan::text::bold
	echo "${*}"
	oscan::text::reset
}

# oscan::text::print_red prints all input in red text
function oscan::text::print_red() {
	oscan::text::red
	echo "${*}"
	oscan::text::reset
}

# oscan::text::print_red_bold prints all input in bold red text
function oscan::text::print_red_bold() {
	oscan::text::red
	oscan::text::bold
	echo "${*}"
	oscan::text::reset
}

# oscan::text::print_green prints all input in green text
function oscan::text::print_green() {
	oscan::text::green
	echo "${*}"
	oscan::text::reset
}

# oscan::text::print_green_bold prints all input in bold green text
function oscan::text::print_green_bold() {
	oscan::text::green
	oscan::text::bold
	echo "${*}"
	oscan::text::reset
}

# oscan::text::print_blue prints all input in blue text
function oscan::text::print_blue() {
	oscan::text::blue
	echo "${*}"
	oscan::text::reset
}

# oscan::text::print_blue_bold prints all input in bold blue text
function oscan::text::print_blue_bold() {
	oscan::text::blue
	oscan::text::bold
	echo "${*}"
	oscan::text::reset
}

# oscan::text::print_yellow prints all input in yellow text
function oscan::text::print_yellow() {
	oscan::text::yellow
	echo "${*}"
	oscan::text::reset
}

# oscan::text::print_yellow_bold prints all input in bold yellow text
function oscan::text::print_yellow_bold() {
	oscan::text::yellow
	oscan::text::bold
	echo "${*}"
	oscan::text::reset
}
