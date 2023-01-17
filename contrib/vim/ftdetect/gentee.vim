fun! s:DetectGentee()
  if getline(1) =~# '^#!.*/usr/local/bin/gentee'
    set filetype=gentee
  endif
endfun

au BufRead,BufNewFile *.g set filetype=gentee
if did_filetype()
  finish
endif
au BufRead,BufNewFile * call s:DetectGentee()


