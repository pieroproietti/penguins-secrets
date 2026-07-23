# penguins-secrets

I am a 65-year-old man who has been working professionally in IT since the second half of the 1980s. During the first half of that decade, out of pure curiosity, I learned programming—and what little English I know—from ZX Spectrum manuals.

The human mind has wonderful potential, but as the years go by, it tends to settle down and learning new things becomes harder. I envy digital natives—those born, say, after 2000—though not too much, because I realize I don't know any of them who are truly "formidable" in this field.

Perhaps it's no coincidence. Ever since the dawn of IT, I've dealt with passwords and authentication issues. Eventually, I had to give in and write passwords down in order to remember them—after all, I was born a human, not a computer.

Perhaps to shirk any real responsibility, a horde of youngsters now decides for me how I must write my passwords for online banking, digital ID, and so on. It was tolerable before the advent of AI: you just created your nice little `secrets.md` file and wrote them down there.

Now, however, AI exists—and even worse, AI Agents. If an agent is malicious, it reads your secrets and robs you *ipso facto*.

We need a better system to protect ourselves: an encrypted volume that stays unmounted most of the time, mounted only on occasion when needed to copy a password, and closed immediately after.

The password is "QuelloCheCazzoViPare" (*"WhateverTheFuckYouWant"*), easily memorable and customizable. It's your responsibility.

Here, I want to type as little as possible: single `s4` CLI in Go with subcommands `s4 create`, `s4 mount`, `s4 umount`, `s4 clone`, and `s4 status`.

## Usage / Uso (`s4`)

```bash
# Build binario e pacchetto .deb
./m
```
# Usage / Utilizzo
s4 create [size]   # Creates, formats LUKS+FAT, and mounts (default size: 64M, e.g. 10M, 64M, 100M)
s4 mount / open    # Unlocks LUKS and mounts FAT32 volume to mnt/
s4 umount / lock   # Unmounts FAT32 AND completely seals LUKS container (use -f to force)
s4 clone [dest]   # Safely backups secrets.img
s4 status         # Displays current mount and LUKS container status
s4 completion     # Generates Bash autocompletion script
```

### Installazione tramite pacchetto `.deb` (Consigliata)
Per una pigrizia davvero professionale, basta installare il pacchetto `.deb`:

```bash
sudo dpkg -i penguins-secrets_1.0.0_amd64.deb
```
Questo installerà in automatico:
- L'eseguibile `s4` in `/usr/bin/s4` (subito disponibile nel `PATH` di sistema)
- L'autocompletamento in `/usr/share/bash-completion/completions/s4` (funzionante subito in Bash)


### Bash Autocompletion & PATH Setup / Configurazione Bash
Aggiungi queste due righe in fondo al tuo file `~/.bashrc`:

```bash
export PATH="$PATH:/home/artisan/penguins-secrets/bin"
source /home/artisan/penguins-secrets/s4-completion.bash
```
oppure esegui direttamente:
```bash
eval "$(s4 completion)"
```


---

### Versione Italiana (Originale)

Sono un uomo di 65 anni, lavoro professionalmente con l'informatica dalla seconda metà degli anni '80 e, durante la prima, per pura curiosità ho appreso la programmazione e, quel poco di inglese che conosco dai manuali dello ZX Spectrum.

La mente umana ha delle splendide possibilità. ma con gli anni tende ad assestarsi ed ad essere più difficile apprendere nuove cose. Invidio i nativi digitali, quelli nati diciamo dopo il 2000, ma non più di tanto perchè mi rendo conto che non conosco alcuno di loro veramente "forte" in questo campo.

Forse non è un caso, di fatto dagli albori della IT ad adesso ho bazzicato con password e problemi di autentificazione, alla fine ho dovuto desistere e decidere di scrivere le password per poterle ricordare ma sono nato uomo non elaboratore.

Forse per scaricarsi da qualsiasi responsabilità, un frotta di giovinastri decide per me come debba scrivere la password per la banca, per lo spid e cosi via. Era pure sottortabile prima dell'avvento della AI, ti facevi il tuo bel file secrets.md e le scrivevi la.

Ora però c'è l'AI e peggio gli agenti AI, se un agente è malizioso legge il tuo secrets e ti rapina ipso facto.

Serve un sistema migliore per proteggersi, un volume criptato che sia sempre smontato e che si monta all'occasione, si copia la password e si richiude subito dopo.

La password è "QuelloCheCazzoViPare" ricordabile a mente e cambiabile pure. è responsabilità vostra.

Qua vorrei scrivere il meno possibile: un'unica utility `s4` in Go con sotto-comandi `s4 create`, `s4 mount`, `s4 umount`, `s4 clone` e `s4 status`.
