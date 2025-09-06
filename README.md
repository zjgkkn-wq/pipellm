# PipeLLM

A simple CLI tool for working with the ChatGPT API using named prompts
and shell aliases.

## Installation

1. Build the binary:
```bash
go build -o pipellm
````

2. Create a config file `~/.pipellm.yaml`:

```yaml
api_key: your_openai_api_key_here

prompts:
- name: summary
  prompt: Summarize this text and provide an overview:

- name: kharms
  prompt: Rewrite this text as if you were Daniil Kharms:
```

3. Generate shell aliases and add them to `.bashrc`:

```bash
./pipellm --bash-alias >> ~/.bashrc
source ~/.bashrc
```

## Usage

Run prompts directly in pipelines:

```bash
echo "Long text" | summary
echo "Plain text" | kharms
```

Or combine them:

```bash
cat error.txt | grep ERROR | summary  | kharms 
Everything has vanished like smoke, the file exists no more.
```

Usage:

```bash
cat dsu.cc | review | summary | kharms
Code comments every day
Inspire me to action
Fix mistakes,
Refine the code,
Clean up with labels and concepts.
Keep it up, developer, in the same spirit!
```
