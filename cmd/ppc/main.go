package main

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    "github.com/7109jun/plusplus-c/diagnostics"
    "github.com/7109jun/plusplus-c/ir"
    "github.com/7109jun/plusplus-c/lexer"
    "github.com/7109jun/plusplus-c/parser"
    "github.com/7109jun/plusplus-c/semantic"
    "github.com/7109jun/plusplus-c/target/x86bios"
)

const version = "1.0.0"

func main() {
    if len(os.Args) < 2 {
        printUsage()
        return
    }

    switch os.Args[1] {
    case "build":
        if len(os.Args) != 3 {
            printUsage()
            os.Exit(1)
        }
        if err := runBuild(os.Args[2]); err != nil {
            fmt.Fprintf(os.Stderr, "error: %v\n", err)
            os.Exit(1)
        }
    case "asm":
        if len(os.Args) != 3 {
            printUsage()
            os.Exit(1)
        }
        if err := runAsm(os.Args[2]); err != nil {
            fmt.Fprintf(os.Stderr, "error: %v\n", err)
            os.Exit(1)
        }
    case "check":
        if len(os.Args) != 3 {
            printUsage()
            os.Exit(1)
        }
        if err := runCheck(os.Args[2]); err != nil {
            fmt.Fprintf(os.Stderr, "error: %v\n", err)
            os.Exit(1)
        }
    case "run":
        if len(os.Args) != 3 {
            printUsage()
            os.Exit(1)
        }
        if err := runProgram(os.Args[2]); err != nil {
            fmt.Fprintf(os.Stderr, "error: %v\n", err)
            os.Exit(1)
        }
    case "version":
        fmt.Printf("++C compiler version %s\n", version)
    case "help", "-h", "--help":
        printUsage()
    default:
        fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
        printUsage()
        os.Exit(1)
    }
}

func printUsage() {
    fmt.Println("usage: ppc <command> <file>")
    fmt.Println("commands:")
    fmt.Println("  build   compile a .ppc file into a BIOS executable image")
    fmt.Println("  asm     emit x86 BIOS assembly")
    fmt.Println("  check   validate syntax and semantics")
    fmt.Println("  run     compile and run in QEMU")
    fmt.Println("  version print compiler version")
}

func runBuild(inputPath string) error {
    src, err := os.ReadFile(inputPath)
    if err != nil {
        return fmt.Errorf("read input: %w", err)
    }

    toks, diags := lexer.Lex(string(src), inputPath)
    if len(diags) > 0 {
        for _, d := range diags {
            fmt.Println(d.String())
        }
    }

    prog, err := parser.ParseProgram(toks, inputPath)
    if err != nil {
        return err
    }
    if err := semantic.Analyze(prog); err != nil {
        return err
    }

    irProg, err := ir.Build(prog)
    if err != nil {
        return err
    }

    art, err := x86bios.Generate(irProg)
    if err != nil {
        return err
    }

    outPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath))
    if err := os.WriteFile(outPath, art.Binary, 0o644); err != nil {
        return fmt.Errorf("write output binary: %w", err)
    }
    fmt.Printf("Built BIOS executable: %s\n", outPath)
    return nil
}

func runAsm(inputPath string) error {
    src, err := os.ReadFile(inputPath)
    if err != nil {
        return err
    }
    toks, diags := lexer.Lex(string(src), inputPath)
    if len(diags) > 0 {
        for _, d := range diags {
            fmt.Println(d.String())
        }
    }
    prog, err := parser.ParseProgram(toks, inputPath)
    if err != nil {
        return err
    }
    if err := semantic.Analyze(prog); err != nil {
        return err
    }
    irProg, err := ir.Build(prog)
    if err != nil {
        return err
    }
    art, err := x86bios.Generate(irProg)
    if err != nil {
        return err
    }
    outPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".asm"
    if err := os.WriteFile(outPath, []byte(art.Assembly), 0o644); err != nil {
        return fmt.Errorf("write asm: %w", err)
    }
    fmt.Printf("Generated assembly: %s\n", outPath)
    return nil
}

func runCheck(inputPath string) error {
    src, err := os.ReadFile(inputPath)
    if err != nil {
        return err
    }
    toks, diags := lexer.Lex(string(src), inputPath)
    if len(diags) > 0 {
        for _, d := range diags {
            fmt.Println(d.String())
        }
    }
    prog, err := parser.ParseProgram(toks, inputPath)
    if err != nil {
        return err
    }
    if err := semantic.Analyze(prog); err != nil {
        return err
    }
    fmt.Printf("Check OK: %s\n", inputPath)
    return nil
}

func runProgram(inputPath string) error {
    if err := runBuild(inputPath); err != nil {
        return err
    }
    bin := strings.TrimSuffix(inputPath, filepath.Ext(inputPath))
    qemu, err := exec.LookPath("qemu-system-i386")
    if err != nil {
        return fmt.Errorf("qemu-system-i386 not found: %w", err)
    }
    cmd := exec.Command(qemu, "-display", "none", "-serial", "stdio", "-drive", "file="+bin+",format=raw,if=ide")
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}
