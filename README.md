# PipeLLM

A simple CLI tool for working with the Gemini API using named prompts
and shell aliases.

## Installation

### Go build

```bash
go build -o pipellm
```

### Create a config file `~/.pipellm.yaml`

```yaml
api_key: your_gemini_api_key_here

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

```bash
kubectl get po -A | summary | kharms
The little pods, they hummed and whirred. Most of them,
you see, were quite content. They resided in grand houses
named `authentik`, `cert-manager`, `kube-system`, and
`zuul`. They performed their duties with a cheerful,
if often inexplicable, vigor.

However, some pods were less enthusiastic about the whole
business. The `vault-server-0` pod, for instance, seemed
to be contemplating its very existence, not quite ready
to engage with the day's proceedings. And a few of the
`ingress-nginx` admission pods, after a brief flurry of
activity, had decided their tasks were, for the moment,
entirely complete. They had finished. They had done. Now,
what? The others, of course, continued their busy, humming
existence, none the wiser.
```
