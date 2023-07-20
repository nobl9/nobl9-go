/*
  Formatter which works on cspell config file and:
  - Sorts the 'words' list.
  - Removes duplicates from 'words' list.
*/

import YAML from 'yaml';
import { readFileSync, writeFileSync } from 'fs';

const CSPELL_CONFIG = "cspell.yaml"

function format() {
  const f = readFileSync(CSPELL_CONFIG, 'utf8')
  const yaml = YAML.parse(f)

  let words = yaml['words']
  words.sort()
  words = [...new Set(words)]
  yaml['words'] = words

  writeFileSync(CSPELL_CONFIG, YAML.stringify(yaml))
}

try { format() } catch (err) { console.error(err) }
