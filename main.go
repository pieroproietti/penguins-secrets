package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config holds paths and names for the encrypted volume setup
type Config struct {
	CryptDir    string
	CryptFile   string
	CryptMnt    string
	CryptMapped string
}

func getConfig() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("failed to get user home directory: %w", err)
	}

	cryptDir := filepath.Join(home, "penguins-secrets")
	return Config{
		CryptDir:    cryptDir,
		CryptFile:   filepath.Join(cryptDir, "secrets.img"),
		CryptMnt:    filepath.Join(cryptDir, "mnt"),
		CryptMapped: "secrets_mapped",
	}, nil
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func isMounted(targetPath string) bool {
	cmd := exec.Command("mountpoint", "-q", targetPath)
	return cmd.Run() == nil
}

func isLuksOpen(mappedName string) bool {
	devPath := filepath.Join("/dev/mapper", mappedName)
	_, err := os.Stat(devPath)
	return err == nil
}

func getCurrentUIDGID() (string, string, error) {
	u, err := user.Current()
	if err != nil {
		return "", "", err
	}
	return u.Uid, u.Gid, nil
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg, err := getConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERRORE] %v\n", err)
		os.Exit(1)
	}

	subcommand := strings.ToLower(os.Args[1])

	switch subcommand {
	case "create":
		size := "64M"
		if len(os.Args) >= 3 {
			size = os.Args[2]
		}
		if err := cmdCreate(cfg, size); err != nil {
			fmt.Fprintf(os.Stderr, "\n[ERRORE] Creazione fallita: %v\n", err)
			os.Exit(1)
		}
	case "mount", "open", "unlock":
		if err := cmdMount(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "\n[ERRORE] Montaggio fallito: %v\n", err)
			os.Exit(1)
		}
	case "umount", "unmount", "close", "lock", "seal":
		force := false
		for _, arg := range os.Args[2:] {
			if arg == "-f" || arg == "--force" {
				force = true
				break
			}
		}
		if err := cmdUmount(cfg, force); err != nil {
			fmt.Fprintf(os.Stderr, "\n[ERRORE] Smontaggio/Chiusura fallita: %v\n", err)
			os.Exit(1)
		}
	case "clone", "backup":
		dest := ""
		if len(os.Args) >= 3 {
			dest = os.Args[2]
		}
		if err := cmdClone(cfg, dest); err != nil {
			fmt.Fprintf(os.Stderr, "\n[ERRORE] Clonazione fallita: %v\n", err)
			os.Exit(1)
		}
	case "status":
		cmdStatus(cfg)
	case "completion":
		cmdCompletion()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "[ERRORE] Comando sconosciuto: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Uso: s4 <comando> [opzioni]")
	fmt.Println()
	fmt.Println("Comandi disponibili:")
	fmt.Println("  create [dimensione]  Crea, formatta e monta un nuovo volume cifrato LUKS/FAT (default: 64M)")
	fmt.Println("  mount | open         Sblocca il volume LUKS e lo monta in mnt/")
	fmt.Println("  umount | lock | seal Smonta il volume FAT32 E chiude completamente il contenitore LUKS (-f per forzare)")
	fmt.Println("  clone [destinazione] Esegue un backup sicuro del volume cifrato secrets.img")
	fmt.Println("  status               Mostra lo stato attuale del volume e del mapper LUKS")
	fmt.Println("  completion [bash]    Genera lo script per l'autocompletamento in Bash")
	fmt.Println("  help                 Mostra questa guida")
}

func cmdCompletion() {
	bashScript := `_s4_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="create mount open unlock umount unmount close lock seal clone backup status completion help"

    if [ $COMP_CWORD -eq 1 ]; then
        COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
        return 0
    fi
}
complete -F _s4_completion s4
`
	fmt.Print(bashScript)
}


func cmdCreate(cfg Config, size string) error {
	fmt.Println("-> Inizializzazione di un nuovo contenitore cifrato...")

	// Se un contenitore o mapper precedente è ancora aperto o montato, lo chiudiamo prima
	if isMounted(cfg.CryptMnt) || isLuksOpen(cfg.CryptMapped) {
		fmt.Println("-> Rilevato un mapper LUKS o mount point residuo attivo. Chiusura preventiva in corso...")
		if err := cmdUmount(cfg, true); err != nil {
			return fmt.Errorf("impossibile chiudere il volume residuo prima della creazione: %w", err)
		}
	}

	if err := os.MkdirAll(cfg.CryptDir, 0755); err != nil {
		return err
	}

	fmt.Printf("-> Allocazione di un file da %s in %s...\n", size, cfg.CryptFile)
	if err := runCommand("fallocate", "-l", size, cfg.CryptFile); err != nil {
		return fmt.Errorf("fallocate fallito: %w", err)
	}

	fmt.Println("-> Formattazione LUKS del contenitore (verrà chiesta la passphrase)...")
	if err := runCommand("sudo", "cryptsetup", "luksFormat", cfg.CryptFile); err != nil {
		return fmt.Errorf("luksFormat fallito: %w", err)
	}

	fmt.Println("-> Apertura del contenitore virtuale...")
	if err := runCommand("sudo", "cryptsetup", "luksOpen", cfg.CryptFile, cfg.CryptMapped); err != nil {
		return fmt.Errorf("luksOpen fallito: %w", err)
	}

	mappedPath := filepath.Join("/dev/mapper", cfg.CryptMapped)
	fmt.Println("-> Formattazione del volume in FAT32...")
	if err := runCommand("sudo", "mkfs.vfat", "-F", "32", mappedPath); err != nil {
		if err2 := runCommand("sudo", "mkfs.vfat", mappedPath); err2 != nil {
			return fmt.Errorf("mkfs.vfat fallito: %w", err)
		}
	}
	fmt.Println("   Formattazione FAT32 completata con successo.")

	if err := os.MkdirAll(cfg.CryptMnt, 0755); err != nil {
		return err
	}

	fmt.Println("-> Montaggio del volume...")
	if err := mountVolume(mappedPath, cfg.CryptMnt); err != nil {
		return err
	}

	fmt.Printf("-> Volume creato e montato con successo in %s!\n", cfg.CryptMnt)
	return nil
}

func getFilesystemType(devPath string) string {
	out, err := exec.Command("sudo", "blkid", "-s", "TYPE", "-o", "value", devPath).Output()
	if err == nil && len(out) > 0 {
		return strings.TrimSpace(string(out))
	}
	out2, err2 := exec.Command("lsblk", "-n", "-o", "FSTYPE", devPath).Output()
	if err2 == nil {
		return strings.TrimSpace(string(out2))
	}
	return ""
}

func cmdMount(cfg Config) error {
	if _, err := os.Stat(cfg.CryptFile); os.IsNotExist(err) {
		return fmt.Errorf("il file %s non esiste. Esegui prima 's4 create'", cfg.CryptFile)
	}

	// Se sia il mapper LUKS che il mount point sono già attivi, il volume è già in uso
	if isMounted(cfg.CryptMnt) && isLuksOpen(cfg.CryptMapped) {
		fmt.Printf("-> Il volume risulta già aperto e montato in %s.\n", cfg.CryptMnt)
		return nil
	}

	// Se il contenitore LUKS è rimasto aperto ma non è montato (es. a causa di uno smontaggio manuale
	// o di un'interruzione precedente), lo chiudiamo prima per garantire che venga sempre richiesta la passphrase.
	if isLuksOpen(cfg.CryptMapped) {
		fmt.Println("-> Rilevato contenitore LUKS aperto ma non montato (stato residuo non sigillato).")
		fmt.Println("-> Chiusura di sicurezza del mapper residuo...")
		if err := runCommand("sudo", "cryptsetup", "luksClose", cfg.CryptMapped); err != nil {
			return fmt.Errorf("impossibile chiudere il contenitore LUKS residuo: %w", err)
		}
	}

	fmt.Println("-> Apertura del contenitore cifrato...")
	if err := runCommand("sudo", "cryptsetup", "luksOpen", cfg.CryptFile, cfg.CryptMapped); err != nil {
		return fmt.Errorf("luksOpen fallito: %w", err)
	}

	if err := os.MkdirAll(cfg.CryptMnt, 0755); err != nil {
		return err
	}

	mappedPath := filepath.Join("/dev/mapper", cfg.CryptMapped)

	// Controllo se il contenitore sbloccato contiene un filesystem valido
	fsType := getFilesystemType(mappedPath)
	if fsType == "" {
		fmt.Printf("-> ATTENZIONE: Il volume cifrato %s è privo di filesystem (non formattato).\n", mappedPath)
		fmt.Println("-> Formattazione automatica in FAT32 in corso...")
		if err := runCommand("sudo", "mkfs.vfat", "-F", "32", mappedPath); err != nil {
			if err2 := runCommand("sudo", "mkfs.vfat", mappedPath); err2 != nil {
				return fmt.Errorf("formattazione FAT32 fallita per %s: %w", mappedPath, err)
			}
		}
		fmt.Println("   Formattazione FAT32 completata con successo.")
	} else {
		fmt.Printf("-> Filesystem rilevato: %s\n", fsType)
	}

	fmt.Println("-> Montaggio del volume...")
	if err := mountVolume(mappedPath, cfg.CryptMnt); err != nil {
		return err
	}

	fmt.Printf("-> Volume aperto e pronto all'uso in %s!\n", cfg.CryptMnt)
	return nil
}

func mountVolume(mappedPath, targetMnt string) error {
	uid, gid, err := getCurrentUIDGID()
	if err != nil {
		return err
	}

	optsVfat := fmt.Sprintf("uid=%s,gid=%s,utf8", uid, gid)
	// 1. Prova montaggio esplicito vfat con opzioni utente (consigliato per FAT32)
	if err := runCommand("sudo", "mount", "-t", "vfat", "-o", optsVfat, mappedPath, targetMnt); err == nil {
		return nil
	}

	optsBasic := fmt.Sprintf("uid=%s,gid=%s", uid, gid)
	// 2. Prova montaggio generico con uid/gid
	if err := runCommand("sudo", "mount", "-o", optsBasic, mappedPath, targetMnt); err == nil {
		return nil
	}

	// 3. Fallback: montaggio semplice (per filesystem che gestiscono direttamente i permessi POSIX come ext4)
	if err := runCommand("sudo", "mount", mappedPath, targetMnt); err == nil {
		return nil
	}

	return fmt.Errorf("impossibile montare il volume %s su %s", mappedPath, targetMnt)
}

func cmdUmount(cfg Config, force bool) error {
	fmt.Println("-> Inizio procedura di smontaggio e sigillatura del volume...")

	if isMounted(cfg.CryptMnt) {
		fmt.Printf("-> [1/2] Smontaggio del filesystem in %s...\n", cfg.CryptMnt)
		err := runCommand("sudo", "umount", cfg.CryptMnt)
		if err != nil && force {
			fmt.Println("   Smontaggio normale fallito. Tentativo di smontaggio forzato (lazy umount -l)...")
			err = runCommand("sudo", "umount", "-l", cfg.CryptMnt)
		}
		if err != nil {
			return fmt.Errorf("umount fallito: %w", err)
		}
		fmt.Println("   FileSystem smontato con successo.")
		time.Sleep(300 * time.Millisecond)
	} else {
		fmt.Printf("-> [1/2] Il volume non risulta montato in %s.\n", cfg.CryptMnt)
	}

	if isLuksOpen(cfg.CryptMapped) {
		fmt.Printf("-> [2/2] Chiusura completa del contenitore LUKS (/dev/mapper/%s)...\n", cfg.CryptMapped)
		var luksErr error
		for attempt := 1; attempt <= 3; attempt++ {
			luksErr = runCommand("sudo", "cryptsetup", "luksClose", cfg.CryptMapped)
			if luksErr == nil {
				break
			}
			time.Sleep(400 * time.Millisecond)
		}

		if luksErr != nil && force {
			fmt.Println("   Tentativo di chiusura forzata deferred del mapper LUKS...")
			_ = runCommand("sudo", "cryptsetup", "luksClose", "--deferred", cfg.CryptMapped)
		}

		if isLuksOpen(cfg.CryptMapped) {
			return fmt.Errorf("luksClose fallito per /dev/mapper/%s: riprova chiudendo terminali o file manager aperti sul volume, oppure usa 's4 umount -f'", cfg.CryptMapped)
		}
		fmt.Println("   Contenitore LUKS chiuso con successo.")
		fmt.Println("-> Volume sigillato in sicurezza!")
	} else {
		fmt.Printf("-> [2/2] Il contenitore LUKS (/dev/mapper/%s) risulta già chiuso.\n", cfg.CryptMapped)
		fmt.Println("-> Volume già sigillato.")
	}

	return nil
}

func cmdClone(cfg Config, dest string) error {
	if _, err := os.Stat(cfg.CryptFile); os.IsNotExist(err) {
		return fmt.Errorf("il file %s non esiste. Niente da clonare", cfg.CryptFile)
	}

	if dest == "" {
		backupDir := filepath.Join(cfg.CryptDir, "backups")
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			return err
		}
		ts := time.Now().Format("20060102_150405")
		dest = filepath.Join(backupDir, fmt.Sprintf("secrets_%s.img", ts))
	}

	wasOpen := isLuksOpen(cfg.CryptMapped) || isMounted(cfg.CryptMnt)
	if wasOpen {
		fmt.Println("-> Il volume è attualmente aperto. Chiusura temporanea prima del backup...")
		if err := cmdUmount(cfg, false); err != nil {
			return fmt.Errorf("impossibile chiudere il volume prima del backup: %w", err)
		}
	}

	fmt.Printf("-> Clonazione di %s in %s ...\n", cfg.CryptFile, dest)
	if err := copyFile(cfg.CryptFile, dest); err != nil {
		return fmt.Errorf("copia fallita: %w", err)
	}
	fmt.Printf("-> Backup completato con successo: %s\n", dest)

	if wasOpen {
		fmt.Println("-> Riapertura del volume...")
		if err := cmdMount(cfg); err != nil {
			return fmt.Errorf("impossibile riaprire il volume post-backup: %w", err)
		}
	}

	return nil
}

func cmdStatus(cfg Config) {
	fmt.Println("--- Stato Penguins-Secrets ---")
	fmt.Printf("File cifrato:   %s\n", cfg.CryptFile)
	if _, err := os.Stat(cfg.CryptFile); os.IsNotExist(err) {
		fmt.Println("Esistenza file: NON PRESENTE")
	} else {
		fmt.Println("Esistenza file: PRESENTE")
	}

	if isLuksOpen(cfg.CryptMapped) {
		fmt.Printf("Contenitore LUKS: APERTO (/dev/mapper/%s)\n", cfg.CryptMapped)
	} else {
		fmt.Println("Contenitore LUKS: CHIUSO")
	}

	if isMounted(cfg.CryptMnt) {
		fmt.Printf("Punto di mount:  MONTATO (%s)\n", cfg.CryptMnt)
	} else {
		fmt.Printf("Punto di mount:  SMONTATO (%s)\n", cfg.CryptMnt)
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Sync()
}

// strconv helper for potential unused check
var _ = strconv.Itoa
