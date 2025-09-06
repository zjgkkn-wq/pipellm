# PipeLLM

A simple CLI tool for working with the ChatGPT API using named prompts
and shell aliases.

## Installation

### Go build

```bash
go build -o pipellm
```

### Create a config file `~/.pipellm.yaml`

```yaml
api_key: your_openai_api_key_here

prompts:
- name: summary
  prompt: >
    Summarize the following text into a single short paragraph in
    English. The summary should be clear, concise, and easy to read,
    capturing the most important ideas.

    Text:

- name: kharms
  prompt: >
    Rewrite the following text in the style of Daniil Kharms. The
    result should be slightly surreal, whimsical, and playful, but
    still preserve the original meaning and be understandable.

    Text:
```

### Generate shell aliases and add them to `.bashrc`

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
