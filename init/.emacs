;; load emacs 24's package system. Add MELPA repository.
(package-initialize)

(when (>= emacs-major-version 24)
  (require 'package)
  (add-to-list
   'package-archives
   '("melpa" . "http://melpa.milkbox.net/packages/")
   t))
(custom-set-variables
 '(package-selected-packages (quote (go-mode))))
(custom-set-faces
 )
(add-hook 'before-save-hook 'gofmt-before-save)

