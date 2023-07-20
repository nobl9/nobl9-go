/*
j Linter which checks things related to maintaining alphabetical order of words in cspell.json.
*/

import { path } from 'path'

const cspellConfigConfigPath = path.resolve(__dirname, '../../cspell.json')
const cspellConfig = require(cspellConfigConfigPath)
const wordList = {
  list: cspellConfig.words,
  filePath: cspellConfigConfigPath
}

const loadedWordList = wordList.list
const expectedWordList = wordList.list.slice().sort((a, b) => a.localeCompare(b))
if (loadedWordList.length !== expectedWordList.length) {
  console.error(`Unexpected error occurred, check source code: ${__filename}`)
  process.exit(2)
}

for (let i = 0; i < loadedWordList.length; i++) {
  const actualWord = loadedWordList[i]
  const expectedWord = expectedWordList[i]
  if (actualWord !== expectedWord) {
    console.error(`
Alphabetical order of words is not maintained in cspell.json
First mismatch:
  actual: ${actualWord}
  expected: ${expectedWord}
`)
    process.exit(1)
  }
}
