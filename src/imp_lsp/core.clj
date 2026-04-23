(ns imp-lsp.core
  (:import
   [io.github.treesitter.jtreesitter Language Parser]
   [java.lang.foreign Arena SymbolLookup])
  (:gen-class))

(defn -main
  "I don't do a whole lot ... yet."
  [& args]
  (println "Hello, World!"))

(defn load-lang
  []
  (let [langpath (concat "/usr/local/lib" (System/mapLibraryName "tree-sitter-imp"))
        arena (Arena/global)
        symbols (SymbolLookup/libraryLookup langpath arena)
        language (Language/load symbols)
        parser (Parser. language)]
    (. parser parse (slurp "../../test.imp"))))
