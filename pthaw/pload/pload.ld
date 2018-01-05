/*
Copyright (c) 2011 The Native Client Authors. All rights reserved.
Use of this source code is governed by a BSD-style license that can be
found in the LICENSE file.

This file is based on nacl_bootstrap.x from Native Client:
https://chromium.googlesource.com/native_client/src/native_client/+/287c0fffd7d3b7f38b887324f261ddb724d38f26/src/trusted/service_runtime/linux/nacl_bootstrap.x

Make everything appear normal */
ENTRY(main)

/*
This is the address where the program text starts.
We set this as low as we think we can get away with.
The common settings for sysctl vm.mmap_min_addr range from 4k to 64k.
*/
TEXT_START = 0x10000;

/*
We specify the program headers we want explicitly, to get the layout
exactly right and to give the "reserve" segment p_flags of zero, so
that it gets mapped as PROT_NONE.
*/
PHDRS {
  text PT_LOAD FILEHDR PHDRS;
  data PT_LOAD;
  reserve PT_LOAD FLAGS(0);
  r_debug PT_LOAD;
  note PT_NOTE;
  stack PT_GNU_STACK FLAGS(6);  /* RW, no E */
}

/* Now we lay out the sections across those segments. */
SECTIONS {
  . = TEXT_START + SIZEOF_HEADERS;

/* Here is the program itself. */
  .text : {
    *(.text*)
  } :text
  .rodata : {
    *(.rodata*)
    *(.eh_frame*)
  }
  etext = .;

/*
Adjust the address for the data segment.  We want to adjust up to
the same address within the page on the next page up.
*/
  . = (ALIGN(CONSTANT(MAXPAGESIZE)) -
       ((CONSTANT(MAXPAGESIZE) - .) & (CONSTANT(MAXPAGESIZE) - 1)));
  . = DATA_SEGMENT_ALIGN(CONSTANT(MAXPAGESIZE), CONSTANT(COMMONPAGESIZE));

  .data : {
    *(.data*)
  } :data
  .bss : {
    *(.bss*)
  }

/*
Now we move up to the next p_align increment, and place the dummy
segment there.  The linker emits this segment with the p_vaddr and
p_memsz we want, which reserves the address space.  But the linker
gives it a p_filesz of zero.  We have to edit the phdr after link
time to give it a p_filesz matching its p_memsz.  That way, the
kernel doesn't think we are preallocating a huge amount of memory.
It just maps it from the file, i.e. way off the end of the file,
which is perfect for reserving the address space.
*/
  . = ALIGN(CONSTANT(COMMONPAGESIZE));
  RESERVE_START = .;
  .reserve : {
    . += (RESERVE_TOP > RESERVE_START) ? (RESERVE_TOP - RESERVE_START) : 0;
  } :reserve


/*
These are empty input sections the linker generates.
If we don't discard them, they pollute the flags in the output segment.
*/
  /DISCARD/ : {
    *(.iplt)
    *(.reginfo)
    *(.rel*)
    *(.igot.plt)
    *(.ARM.exidx)
  }
}
