package completion

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/bangmodtechnology/kselect/pkg/registry"
)

// GenerateBash outputs a bash completion script for kselect.
func GenerateBash(w io.Writer) {
	resources := getResourceNames()

	fmt.Fprintf(w, `# bash completion for kselect
# Usage: source <(kselect completion bash)
# Or:    kselect completion bash > /etc/bash_completion.d/kselect

_kselect_completions() {
    local cur prev prev_upper
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    prev_upper=$(echo "$prev" | tr '[:lower:]' '[:upper:]')

    local resources="%s"
    local keywords="FROM WHERE ORDER BY LIMIT OFFSET GROUP HAVING AND OR LIKE IN NOT DISTINCT ASC DESC INNER LEFT RIGHT JOIN ON"
    local operators="GT GE LT LE NE EQ"
    local flags="-n -A -o -watch -no-color -version -list -plugins -interval"
    local formats="table json yaml csv wide"

    case "$prev_upper" in
        FROM|JOIN)
            COMPREPLY=($(compgen -W "$resources" -- "$cur"))
            return ;;
        -O)
            COMPREPLY=($(compgen -W "$formats" -- "$cur"))
            return ;;
        -N)
            local ns
            ns=$(kubectl get namespaces -o jsonpath='{.items[*].metadata.name}' 2>/dev/null)
            COMPREPLY=($(compgen -W "$ns" -- "$cur"))
            return ;;
        ORDER|GROUP)
            COMPREPLY=($(compgen -W "BY" -- "$cur"))
            return ;;
        BY)
            COMPREPLY=($(compgen -W "ASC DESC" -- "$cur"))
            return ;;
        COMPLETION)
            COMPREPLY=($(compgen -W "bash zsh" -- "$cur"))
            return ;;
    esac

    if [[ "$cur" == -* ]]; then
        COMPREPLY=($(compgen -W "$flags" -- "$cur"))
        return
    fi

    COMPREPLY=($(compgen -W "$keywords $operators $resources completion" -- "$cur"))
}

complete -o default -F _kselect_completions kselect
`, strings.Join(resources, " "))
}

// GenerateZsh outputs a zsh completion script for kselect.
func GenerateZsh(w io.Writer) {
	resources := getResourceNames()

	fmt.Fprintf(w, `#compdef kselect
# zsh completion for kselect
# Usage: source <(kselect completion zsh)
# Or:    kselect completion zsh > "${fpath[1]}/_kselect"

_kselect() {
    local -a resources keywords flags formats operators

    resources=(%s)
    keywords=(FROM WHERE ORDER BY LIMIT OFFSET GROUP HAVING AND OR LIKE IN NOT DISTINCT ASC DESC INNER LEFT RIGHT JOIN ON)
    flags=(-n -A -o -watch -no-color -version -list -plugins -interval)
    formats=(table json yaml csv wide)
    operators=(GT GE LT LE NE EQ)

    local prev_word="${words[CURRENT-1]}"
    local prev_upper="${(U)prev_word}"

    case "$prev_upper" in
        FROM|JOIN)
            compadd -a resources
            ;;
        -O)
            compadd -a formats
            ;;
        -N)
            local -a ns
            ns=(${(f)"$(kubectl get namespaces -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' 2>/dev/null)"})
            compadd -a ns
            ;;
        ORDER|GROUP)
            compadd BY
            ;;
        BY)
            compadd ASC DESC
            ;;
        COMPLETION)
            compadd bash zsh
            ;;
        *)
            if [[ "$words[CURRENT]" == -* ]]; then
                compadd -a flags
            else
                compadd -a keywords operators resources
                compadd completion
            fi
            ;;
    esac
}

compdef _kselect kselect
`, strings.Join(resources, " "))
}

func getResourceNames() []string {
	reg := registry.GetGlobalRegistry()
	resources := reg.ListResources()

	var names []string
	seen := make(map[string]bool)
	for _, res := range resources {
		if !seen[res.Name] {
			names = append(names, res.Name)
			seen[res.Name] = true
		}
		for _, alias := range res.Aliases {
			if !seen[alias] {
				names = append(names, alias)
				seen[alias] = true
			}
		}
	}
	sort.Strings(names)
	return names
}
