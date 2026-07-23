# Script di autocompletamento Bash per s4 (penguins-secrets)
# Per attivarlo subito nella sessione corrente:
#   source /home/artisan/penguins-secrets/s4-completion.bash
#
# Per attivarlo in automatico a ogni avvio di terminale, aggiungi a ~/.bashrc:
#   source /home/artisan/penguins-secrets/s4-completion.bash

_s4_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="create mount open umount unmount close clone backup status completion help"

    if [ $COMP_CWORD -eq 1 ]; then
        COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
        return 0
    fi
}
complete -F _s4_completion s4
