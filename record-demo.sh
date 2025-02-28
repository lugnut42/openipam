#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'
BOLD='\033[1m'

echo -e "${BOLD}${GREEN}==================================================${NC}"
echo -e "${BOLD}${GREEN}    OPENIPAM ASCIINEMA RECORDING HELPER${NC}"
echo -e "${BOLD}${GREEN}==================================================${NC}"
echo -e "${CYAN}This script will:${NC}"
echo -e "${CYAN}1. Start an asciinema recording${NC}"
echo -e "${CYAN}2. Run the OpenIPAM demo script${NC}"
echo -e "${CYAN}3. Save the recording to openipam-demo.cast${NC}"
echo -e ""
echo -e "${BOLD}${CYAN}The recording will automatically stop when the demo completes.${NC}"
echo -e "${BOLD}${CYAN}Starting in 3 seconds...${NC}"
sleep 3

# Record with a 2-second idle time cutoff and run the demo script
asciinema rec -i 2 -t "OpenIPAM - IP Address Management Tool Demo" -c "./demo.sh" openipam-demo.cast

echo -e "${BOLD}${GREEN}==================================================${NC}"
echo -e "${BOLD}${GREEN}    RECORDING COMPLETE!${NC}"
echo -e "${BOLD}${GREEN}==================================================${NC}"
echo -e "${CYAN}To upload to asciinema.org:${NC}"
echo -e "${BOLD}${CYAN}  asciinema upload openipam-demo.cast${NC}"
echo -e ""
echo -e "${CYAN}After uploading, update the README.md with your recording ID:${NC}"
echo -e "${BOLD}${CYAN}  1. Copy the URL from the upload (e.g., https://asciinema.org/a/12345)${NC}"
echo -e "${BOLD}${CYAN}  2. Extract the ID number (12345 in the example)${NC}"
echo -e "${BOLD}${CYAN}  3. Replace 'YOUR_ASCIINEMA_ID' in the README.md${NC}"