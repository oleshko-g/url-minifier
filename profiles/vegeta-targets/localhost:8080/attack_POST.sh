#!/bin/zsh
vegeta attack -targets ./POST-/targets.txt -workers 10 -duration 10s -rate 100 -name 'POST /'
