title: Writing a Linux executable by hand
excerpt: I wanted to understand what a compiler actually outputs, so I stopped treating executables like magic blobs and learned how to build one by hand. It turns out they're a lot less mysterious than I thought they'd be.
date: 2026-05-19T21:59:00Z
public: true

---

I've built a handful of toy programming languages over the years, purely for the fun of it. Interpreters are a fun weekend project, and byte-code VMs keep me busy for a couple of weeks. However, the closest I've come to building a compiler is a (very) simple front-end for [LLVM][llvm], which quickly joined my project graveyard because I get burnt out too easily from C++.

Recently, I've been itching to give a compiler another shot, without LLVM. This got me thinking:

I don't really understand executable files.

If I want to build a compiler, I should at least be familiar with what a compiler is supposed to output. All I knew before I started this journey was that executable files are binary, which means that I should be able to learn the binary format and produce executables that conform to it. Theoretically, I should even be able to do so manually, byte-by-byte, without a compiler.

Right?

As it turns out, yes! And honestly, it's shockingly straightforward. This post is a retelling of how I learned what (almost) every byte in an executable means and how to create a working hello world executable without a compiler.

## The target

Before we start slinging bytes at files, we need to talk about the **target**.

Executable files are famously not very portable. They are written for a specific type of OS version and CPU architecture. This is called the target.

Since I run the people's OS on a Ryzen 9 7950X3D CPU, my target will be Linux x86_64.

If you're also on Linux, or have access to Linux through something like [WSL][wsl] or a VM, and you have an x86_64 CPU, then you can follow along and write your own executable too. If you're on a different CPU architecture, you'll need to deviate in a few places. An LLM should be able to help you here.

If you're neither on Linux nor on x86_64: I dunno, enjoy reading I guess.

## Where baby elves come from

Executables on Linux use a format called [ELF][elf], as do shared libraries & relocatable objects. It's very well and openly documented; Wikipedia alone is an amazing reference for the format.

The general structure of an ELF file is as follows:
- one ELF header
- a program header table
- optional section header table
- the binary data referenced by the above tables

The file begins with a section that is called the "ELF header". It provides various details about the file, such as its format, target, and where the data can be found. On 64-bit systems, this header is just 64 bytes.

If you're on Linux, you can use `readelf -h <FILE>` to read the ELF header of an executable and print the information in a readable format.

For example, here's what `readelf` says about `ls`:

```
$ readelf -h /usr/bin/ls
ELF Header:
  Magic:   7f 45 4c 46 02 01 01 00 00 00 00 00 00 00 00 00
  Class:                             ELF64
  Data:                              2's complement, little endian
  Version:                           1 (current)
  OS/ABI:                            UNIX - System V
  ABI Version:                       0
  Type:                              DYN (Position-Independent Executable file)
  Machine:                           Advanced Micro Devices X86-64
  Version:                           0x1
  Entry point address:               0x54f0
  Start of program headers:          64 (bytes into file)
  Start of section headers:          160680 (bytes into file)
  Flags:                             0x0
  Size of this header:               64 (bytes)
  Size of program headers:           56 (bytes)
  Number of program headers:         14
  Size of section headers:           64 (bytes)
  Number of section headers:         28
  Section header string table index: 27
```

The first 16 bytes of the ELF header are known as the identification string (`e_ident`). It starts with the "magic number", which is just `0x7F` followed by the string "ELF" in ASCII: `0x45 0x4c 0x46`.

If you ever need to programmatically check if a file is an ELF file, you only need to check the first 4 bytes!

I won't bore you with a byte-by-byte explanation of the entire header. The [ELF wikipedia page][elf] has [a great table][elf_header] that you can reference, and in fact I encourage you to take a few minutes to skim through it to get an idea of what we'll be doing in the next section. We learn by doing around here, so let the doing begin.

## Let's write some bytes

First, we need a way to write raw binary data directly to a file. A text editor won't work, since that will encode everything as text. We could use a hex editor, but we're programmers; we can write a program to do this for us.

Let's start by writing the ELF header's identification string to a file.

```c
#include <stdio.h>
#include <stdint.h>

int main() {
    FILE *f = fopen("my_elf", "w");

    uint8_t ident[16] = {
        0x7F, 'E', 'L', 'F', // magic string
        2,                   // 64-bit file format
        1,                   // little-endian
        1,                   // ELF version
        0,                   // no particular ABI
        0,                   // ABI version
        0, 0, 0, 0, 0, 0, 0, // 7 bytes of padding
    };
    fwrite(ident, sizeof(uint8_t), sizeof(ident), f);

    fclose(f);
}
```

Next is the object file type, which is two bytes. For executables, this should be a `0x0002`.

```c
#include <stdio.h>
#include <stdint.h>

int main() {
    FILE *f = fopen("my_elf", "w");

    uint8_t ident[16] = { /** omitted for brevity */ };
    fwrite(ident, sizeof(uint8_t), sizeof(ident), f);

    uint16_t type = 0x0002;
    fwrite(&type, sizeof(uint16_t), 1, f);

    fclose(f);
}
```

Now, we could keep going and write the rest of the file like this - and that's what I initially did - but as it turns out, there's a much better way.

I don't recall exactly what I was searching for, but I stumbled upon [this ELF man page][elf_man_page] and was surprised to see it mention `elf.h`.

Is this a real C header file? And is it shipped with Linux?

```
$ find /usr/include -name elf.h
/usr/include/elf.h
/usr/include/asm/elf.h
/usr/include/sys/elf.h
/usr/include/linux/elf.h
```

Yes, and yes. Cool. What's in it?

Oh, nothing much, just **THE FREAKING TYPEDEFS** for all ELF-related things!

```c
typedef struct
{
    unsigned char e_ident[EI_NIDENT]; /* Magic number and other info */
    Elf64_Half    e_type;             /* Object file type */
    Elf64_Half    e_machine;          /* Architecture */
    Elf64_Word    e_version;          /* Object file version */
    Elf64_Addr    e_entry;            /* Entry point virtual address */
    Elf64_Off     e_phoff;            /* Program header table file offset */
    Elf64_Off     e_shoff;            /* Section header table file offset */
    Elf64_Word    e_flags;            /* Processor-specific flags */
    Elf64_Half    e_ehsize;           /* ELF header size in bytes */
    Elf64_Half    e_phentsize;        /* Program header table entry size */
    Elf64_Half    e_phnum;            /* Program header table entry count */
    Elf64_Half    e_shentsize;        /* Section header table entry size */
    Elf64_Half    e_shnum;            /* Section header table entry count */
    Elf64_Half    e_shstrndx;         /* Section header string table index */
} Elf64_Ehdr;
```

It contains struct types for the ELF header (seen above), program headers, constant definitions for the field values, and so much more.

The struct types in particular are incredibly helpful, since we can now instantiate and write a struct, rather than each byte procedurally.

We'll start by creating just an ELF header for now. This should be as easy as importing the header file, creating a struct of type `Elf64_Ehdr`, and assigning a value to every field. The provided constants do a great job of increasing readability.

```c
#include <elf.h>
#include <stdint.h>
#include <stdio.h>

int main() {
    Elf64_Ehdr e_hdr = {
        .e_ident =
            {
                ELFMAG0,       // 0x7f
                ELFMAG1,       // "E"
                ELFMAG2,       // "L"
                ELFMAG3,       // "F"
                ELFCLASS64,    // 64-bit
                ELFDATA2LSB,   // little-endian
                EV_CURRENT,    // elf v1
                ELFOSABI_NONE, // no particular ABI
                0,             // ABI version
            },
        .e_type = ET_EXEC,                 // type = executable file
        .e_machine = EM_X86_64,            // targeting x86_64 processors
        .e_version = EV_CURRENT,           // always 1 (EV_CURRENT)
        .e_entry = 0,                      // [*] memory address of 1st instruction
        .e_phoff = sizeof(Elf64_Ehdr),     // [*] file offset of program header tables
        .e_shoff = 0,                      // file offset of section header tables
        .e_flags = 0,                      // always 0
        .e_ehsize = sizeof(e_hdr),         // size of the ELF header
        .e_phentsize = sizeof(Elf64_Phdr), // size of a program header
        .e_phnum = 2,                      // [*] number of program headers
        .e_shentsize = sizeof(Elf64_Shdr), // size of a section header
        .e_shnum = 0,                      // number of section headers
        .e_shstrndx = SHN_UNDEF,           // not important
    };

    FILE *f = fopen("my_elf", "w");
    fwrite(&e_hdr, sizeof(e_hdr), 1, f);
    fclose(f);
}
```

The comments should help explain most of it, but I've also marked 3 fields with `[*]` to elaborate further:

- The `e_entry` field is `0` for now since we don't yet know the memory address of the first instruction. We will update this later.
- The program header table is next after the ELF header in the file, so the offset `e_phoff` is just the size of a single ELF header. Makes sense?
- The `e_phnum` field is `2` because we will need 2 program headers for our executable. I'll explain why in the next section.

We will also be ignoring section headers for today, since we don't need them. So we keep `e_shoff` and `e_shnum` set to `0`.

And that's the ELF header done!

Believe it or not, this is almost a valid ELF file. Try compiling and running the program to produce a `my_elf` file, then run `readelf -h my_elf`. You should also get a warning about the lack of program headers, since the file claims that it has 2 of them, when in fact it has none. Let's address that next.

## Program headers

Program headers tell the kernel how to set up the memory of the process. They describe a single memory segment, and tell the kernel which part of our binary file contains the data that we want to load into it.

For our hello world executable, we will need 2 memory segments:

1. one that contains the data (the hello world string)
1. one that contains the instructions

After we have written these 2 program headers to the file, we will then be writing the actual data and the instructions. Hopefully now you can start to see how the file will be coming together.

Let's start with the program header for the data memory segment.

```c
#define V_ADDR_BASE 0x400000
#define PAGE_SIZE 0x1000
#define EHDR_SIZE sizeof(Elf64_Ehdr)
#define PHDR_SIZE sizeof(Elf64_Phdr)

unsigned char data[] = "Hello, world!\n";
unsigned long data_len = sizeof(data);
unsigned long data_offset = EHDR_SIZE + (2 * PHDR_SIZE);
unsigned long data_vaddr = V_ADDR_BASE + PAGE_SIZE + data_offset;

Elf64_Phdr data_phdr = {
    .p_type = PT_LOAD,       // type: loadable segment
    .p_flags = PF_R,         // flags: readable
    .p_offset = data_offset, // offset in file where the data is found
    .p_vaddr = data_vaddr,   // virtual address of the memory segment
    .p_paddr = data_vaddr,   // physical address - not important
    .p_filesz = data_len,    // segment size in file
    .p_memsz = data_len,     // segment size in memory
    .p_align = PAGE_SIZE,    // segment alignment
};
```

Again, I hope the comments help make most of the code self-explanatory, with perhaps only a few things needing further clarification:

The `p_align` field tells the kernel what byte alignment this segment expects. The key idea is that the segment should line up byte-wise the same way in the file as it does in memory, so `p_offset` and `p_vaddr` need to agree modulo `p_align`. In normal Linux executables this is usually `0x1000` (4 KiB), the memory page size, because memory is mapped in pages.

The `p_offset` field points to the actual data inside the ELF file. Recall that the file starts with an ELF header, then we'll have 2 program headers, and then the contents of the memory segments. So the `data_offset` is just the size of the ELF header plus 2 program headers.

The `p_vaddr` field is the virtual memory address that we want to assign to this segment. Why `0x400000`? From what I could find, this seems to be a convention. Don't get too hung up on this - it's just a virtual address. What's important is that the data in memory is byte-aligned in the same way as it is in the file. We can easily guarantee this by just adding the page size and file offset.

---

Now let's do the same thing for the instructions' memory segment.

```c
unsigned char instructions[] = {}; // we will populate this in the next section
unsigned long inst_len = sizeof(instructions);
unsigned long inst_offset = data_offset + data_len; // immediately after the data
unsigned long inst_vaddr = V_ADDR_BASE + PAGE_SIZE + inst_offset;

Elf64_Phdr inst_phdr = {
    .p_type = PT_LOAD,       // type = loadable segment
    .p_flags = PF_R | PF_X,  // flags = readable & executable
    .p_offset = inst_offset, // offset in file where the data is found
    .p_vaddr = inst_vaddr,   // virtual address of the memory segment
    .p_paddr = inst_vaddr,   // physical address - not important
    .p_filesz = inst_len,    // segment size in file
    .p_memsz = inst_len,     // segment size in memory
    .p_align = PAGE_SIZE,    // segment alignment
};
```

Since this memory segment will contain instructions, we need to also include the `PF_X` flag to mark it as executable.

Note that this time the offset of the instructions is dependent on the previous memory segment. Since we will be writing the data first and the instructions second, the offset for the instructions needs to point to just after the end of the data, i.e. the data offset plus the data length.

This is also a good time to go back to the ELF header and update the `e_entry` field, since now we know what the memory address of the first instruction is: `inst_vaddr`. This means we will also need to move the ELF header struct _below_ the instructions program header in the code, so that we can reference the `inst_vaddr` variable in the ELF header struct.

## Machine code, beep boop

Now we get to the interesting part! We need to generate the machine code for a hello world program. The easiest way to do this is to start with assembly, and then assemble it.

Here's what a hello world program might look like in assembly, using NASM syntax. I personally find NASM to be the most readable for beginners.

```nasm
; write(stdout, &msg, msg_len)
mov    rax, 1    ; rax = 1 (write)
mov    rdi, 1    ; rdi = 1 (stdout)
movabs rsi, 0    ; rsi = &msg    (0 for now)
mov    rdx, 0    ; rdx = msg_len (0 for now)
syscall

; exit(69)
mov rax, 60      ; rax = 60 (exit)
mov rdi, 69      ; rdi = 69
syscall
```

(Note: the above is not a complete NASM program. It lacks section labels, an entry point, and the message itself. But we're only interested in the instructions here.)

To print hello world, our program needs to tell the kernel to write a string to its [standard output stream][std_streams] (STDOUT). We do this by calling the `write` [system call][syscalls], which you can think of as kernel functions. When the `syscall` instruction is invoked, the kernel will look at the `rax` register to see what syscall we want to invoke. Some of the other registers, like `rdi`, `rsi`, and `rdx`, are used to pass arguments to syscalls.

We also need to explicitly tell the program to stop running with the `exit` syscall, which takes the exit code as argument. If we omit this syscall, the program will attempt to run whatever instruction is found next in memory after the `write` syscall, which may not be an instruction and cause undefined behavior. Always include an `exit`! Or don't hand-roll executables!

There are many references for Linux syscalls. The Chromium project has a very extensive one that you can find [here][linux_syscalls].

As you can see in our program, we call the `write` syscall with 3 arguments: `1` for STDOUT, `0` for the message pointer, and `0` for the message length. We will update the pointer and length later, once we have the compiled machine code.

To turn our assembly into machine code, go to [x86_64 playground](https://app.x64.halb.it/), switch the compiler to NASM (using the menu next to the **Compile** button), paste the above assembly code and click **Compile**.

We're interested in the **Disassembly** section. Copy the second column of each instruction, up until the last `syscall` instruction. You should have something like this:

```
b801000000
bf01000000
48 be0000000000000000
ba00000000
0f 05
b83c000000
bf45000000
0f 05
```

This is the machine code equivalent of the hello world assembly. The next step is splitting the above code into bytes so that we can store it in our instructions array in the source code.

How do we do that? Easy. The above machine code is in hexadecimal. A single byte goes up to 255, which is `ff` in hex. That's 2 hexadecimal characters. So every 2 characters in the machine code is a single byte. And in C, hex literals need to be prefixed with `0x`.

So we end up with this:

```c
unsigned char instructions[] = {
    0xb8, 0x01, 0x00, 0x00, 0x00, // mov    rax, 1 | "write" syscall
    0xbf, 0x01, 0x00, 0x00, 0x00, // mov    rdi, 1 | 1st arg: stdout
    0x48, 0xbe, 0x00, 0x00, 0x00, // movabs rsi, ? | 2nd arg: ptr to msg
    0x00, 0x00, 0x00, 0x00, 0x00, //               |
    0xba, 0x00, 0x00, 0x00, 0x00, // mov rdx, ?    | 3rd arg: msg length
    0x0f, 0x05,                   // syscall       | invoke syscall
    0xb8, 0x3c, 0x00, 0x00, 0x00, // mov rax, 60   | "exit" syscall
    0xbf, 0x45, 0x00, 0x00, 0x00, // mov rdi, 69   | 1st arg: exit code
    0x0f, 0x05,                   // syscall       | invoke syscall
};
```

There's just one thing missing: we need to change the message pointer and length from zeroes to their real value.

Let's identify where they should go. The message pointer is an operand of the `movabs` instruction (`0x48 0xbe`), on the 3rd row. So that's index 12 with a length of 8 bytes (this instruction takes up 2 rows). The message length is the operand of the next instruction, on the 5th row. That's index 21 with a length of 4 bytes.

The message pointer should point to the hello world string in the data memory segment, which has an address of `data_vaddr`. Since the segment only contains the string, the address to the string is also just `data_vaddr`. And we already know the length of the message: `data_len`.

So we just need to insert the values of `data_vaddr` and `data_len` into the instructions array, at indexes 12 and 21 respectively.

Let's define a helper function to make this simple.

```c
/**
 * Inserts `value` as bytes (in little-endian) into the `target` at `idx`. Stops
 * after `num` bytes.
 */
void insert_little_endian(uint64_t value, uint8_t *target, size_t idx, size_t num) {
    for (int i = 0; i < num; i++) {
        uint64_t mask = 0xff << (8 * i);
        uint8_t byte = (value & mask) >> (8 * i);
        target[idx + i] = byte;
    }
}
```

The function may look scary but there's really not much going on here. We iterate `num` times, taking the next smallest byte from `value` and writing it to `target` at an offset of `idx`.

I should also mention that my system's byte order is little-endian. You may recall that we already specified this [when we wrote the ELF header](#let-s-write-some-bytes). These days, you'd be hard-pressed to find a big-endian consumer CPU. But if you want to check your machine, you can run this command: `lscpu | grep 'Byte Order'`.

Anyway, now we can do the following:

```c
unsigned char instructions[] = {
    // instructions omitted for brevity
};
insert_little_endian(data_vaddr, instructions, 12, 8);
insert_little_endian(data_len, instructions, 21, 4);
```

And our machine code instructions are complete!

## Put it all together

The final step is to write everything to a file (in the right order!) and mark that file as executable. Put it all together and you should have something like this:

```c
#include <elf.h>
#include <stdint.h>
#include <stdio.h>
#include <sys/stat.h>

#define V_ADDR_BASE 0x400000
#define PAGE_SIZE 0x1000
#define EHDR_SIZE sizeof(Elf64_Ehdr)
#define PHDR_SIZE sizeof(Elf64_Phdr)
#define SHDR_SIZE sizeof(Elf64_Shdr)

// forward declaration
void insert_little_endian(uint64_t value, uint8_t *target, size_t idx, size_t num);

int main() {
    unsigned char data[] = "Hello, world!\n";
    unsigned long data_len = sizeof(data);
    unsigned long data_offset = EHDR_SIZE + (2 * PHDR_SIZE);
    unsigned long data_vaddr = V_ADDR_BASE + PAGE_SIZE + data_offset;

    Elf64_Phdr data_phdr = {
        .p_type = PT_LOAD,
        .p_flags = PF_R,         // flags = readable
        .p_offset = data_offset, // offset in file where data is found
        .p_vaddr = data_vaddr,   // virtual address of the memory segment
        .p_paddr = data_vaddr,   // physical address - not important
        .p_filesz = data_len,    // segment size in file
        .p_memsz = data_len,     // segment size in memory
        .p_align = PAGE_SIZE,    // segment alignment
    };

    unsigned char instructions[] = {
        0xb8, 0x01, 0x00, 0x00, 0x00, // mov    rax, 1 | "write" syscall
        0xbf, 0x01, 0x00, 0x00, 0x00, // mov    rdi, 1 | 1st arg: stdout
        0x48, 0xbe, 0x00, 0x00, 0x00, // movabs rsi, ? | 2nd arg: ptr to msg
        0x00, 0x00, 0x00, 0x00, 0x00, //               |
        0xba, 0x00, 0x00, 0x00, 0x00, // mov rdx, ?    | 3rd arg: msg length
        0x0f, 0x05,                   // syscall       | invoke syscall
        0xb8, 0x3c, 0x00, 0x00, 0x00, // mov rax, 60   | "exit" syscall
        0xbf, 0x45, 0x00, 0x00, 0x00, // mov rdi, 69   | 1st arg: exit code
        0x0f, 0x05,                   // syscall       | invoke syscall
    };

    insert_little_endian(data_vaddr, instructions, 12, 8);
    insert_little_endian(data_len, instructions, 21, 4);

    unsigned long inst_len = sizeof(instructions);
    unsigned long inst_offset = data_offset + data_len;
    unsigned long inst_vaddr = V_ADDR_BASE + PAGE_SIZE + inst_offset;

    Elf64_Phdr inst_phdr = {
        .p_type = PT_LOAD,       // type = loadable segment
        .p_flags = PF_R | PF_X,  // flags = readable & executable
        .p_offset = inst_offset, // offset in file where segment data is found
        .p_vaddr = inst_vaddr,   // The virtual address of the memory segment
        .p_paddr = inst_vaddr,   // The physical address - not important
        .p_filesz = inst_len,    // Segment size in file
        .p_memsz = inst_len,     // Segment size in memory
        .p_align = PAGE_SIZE,    // Segment alignment
    };

    Elf64_Ehdr e_hdr = {
        .e_ident =
            {
                ELFMAG0,       // 0x7f
                ELFMAG1,       // "E"
                ELFMAG2,       // "L"
                ELFMAG3,       // "F"
                ELFCLASS64,    // 64-bit
                ELFDATA2LSB,   // little-endian
                EV_CURRENT,    // elf v1
                ELFOSABI_NONE, // no particular ABI
                0,             // ABI version
            },
        .e_type = ET_EXEC,                 // type = executable file
        .e_machine = EM_X86_64,            // targeting x86_64 processors
        .e_version = EV_CURRENT,           // always 1 or EV_CURRENT
        .e_entry = inst_vaddr,             // address of first instruction
        .e_phoff = EHDR_SIZE,              // offset of program headers
        .e_shoff = 0,                      // offset of section headers
        .e_flags = 0,                      // always 0
        .e_ehsize = EHDR_SIZE,             // size of the ELF header
        .e_phentsize = PHDR_SIZE,          // program header table entry size
        .e_phnum = 2,                      // number of program headers
        .e_shentsize = SHDR_SIZE,          // section header table entry size
        .e_shnum = 0,                      // number of section header table entries
        .e_shstrndx = SHN_UNDEF,           // not important
    };

    FILE *f = fopen("my_elf", "w");
    fwrite(&e_hdr, EHDR_SIZE, 1, f);       // write ELF header
    fwrite(&data_phdr, PHDR_SIZE, 1, f);   // write data prog header
    fwrite(&inst_phdr, PHDR_SIZE, 1, f);   // write inst prog header
    fwrite(data, data_len, 1, f);          // write data
    fwrite(instructions, inst_len, 1, f);  // write instructions
    fchmod(fileno(f), 0755);
    fclose(f);
}

void insert_little_endian(uint64_t value, uint8_t *target, size_t idx, size_t num) {
    for (int i = 0; i < num; i++) {
        uint64_t mask = 0xff << (8 * i);
        uint8_t byte = (value & mask) >> (8 * i);
        target[idx + i] = byte;
    }
}
```

Now all that's left is to compile this bad boy and run it to produce our custom ELF file.

```
$ gcc -o elfmaker main.c
$ ./elfmaker
$ ./my_elf
Hello, world!
```

And if you check the exit code:

```
$ echo $?
69
```

Nice.

If you've been following along:

**Congratulations**! You just built a compiler in less than a hundred lines!

True, it's the world's most pointless compiler, and if we're being pedantic it's technically just a code generator, since we're not translating a source language. But hey, now you have something you can build upon. Here's an exercise for you: swap out the hard-coded hello world string with a string that you read from a text file. That's sort of a language, right?

By the way, if you want to admire your handiwork, run `hexdump -C my_elf`.

```
$ hexdump -C my_elf
00000000  7f 45 4c 46 02 01 01 00  00 00 00 00 00 00 00 00  |.ELF............|
00000010  02 00 3e 00 01 00 00 00  bf 10 40 00 00 00 00 00  |..>.......@.....|
00000020  40 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |@...............|
00000030  00 00 00 00 40 00 38 00  02 00 40 00 00 00 00 00  |....@.8...@.....|
00000040  01 00 00 00 04 00 00 00  b0 00 00 00 00 00 00 00  |................|
00000050  b0 10 40 00 00 00 00 00  b0 10 40 00 00 00 00 00  |..@.......@.....|
00000060  0f 00 00 00 00 00 00 00  0f 00 00 00 00 00 00 00  |................|
00000070  00 10 00 00 00 00 00 00  01 00 00 00 05 00 00 00  |................|
00000080  bf 00 00 00 00 00 00 00  bf 10 40 00 00 00 00 00  |..........@.....|
00000090  bf 10 40 00 00 00 00 00  27 00 00 00 00 00 00 00  |..@.....'.......|
000000a0  27 00 00 00 00 00 00 00  00 10 00 00 00 00 00 00  |'...............|
000000b0  48 65 6c 6c 6f 2c 20 77  6f 72 6c 64 21 0a 00 b8  |Hello, world!...|
000000c0  01 00 00 00 bf 01 00 00  00 48 be b0 10 40 00 00  |.........H...@..|
000000d0  00 00 00 ba 0f 00 00 00  0f 05 b8 3c 00 00 00 bf  |...........<....|
000000e0  45 00 00 00 0f 05                                 |E.....|
000000e6
```

That, is a 230-byte hand-rolled Linux x86_64 executable. Everything the kernel needs to spawn a process. And it's beautiful.

## Gotta Go

When I had finished the first working version of this program, it took me a few hours of playing around with it to realize that, in my pursuit of learning more about compilers, I ended up building one.

But now it was time for me to move on from C. Don't get me wrong, I like C. C is simple, and plenty powerful. Who doesn't like a good sawed-off shotgun to blow your feet clean off?

These days, Go is my mistress. And as luck would have it, Go has an [elf package](https://pkg.go.dev/debug/elf) that is essentially the Go equivalent of `elf.h`. So that's where I'll be continuing this little compiler adventure of mine.

Anyway, hope you learned something. Happy hacking :)

[llvm]: https://llvm.org/
[elf]: https://en.wikipedia.org/wiki/Executable_and_Linkable_Format
[wsl]: https://learn.microsoft.com/en-us/windows/wsl/install
[elf_header]: https://en.wikipedia.org/wiki/Executable_and_Linkable_Format#ELF_header
[elf_man_page]: https://www.man7.org/linux/man-pages/man5/elf.5.html
[abi]: https://en.wikipedia.org/wiki/Application_binary_interface
[syscalls]: https://en.wikipedia.org/wiki/System_call
[linux_syscalls]: https://www.chromium.org/chromium-os/developer-library/reference/linux-constants/syscalls
[std_streams]: https://en.wikipedia.org/wiki/Standard_streams
