open-expo-port:
  #!/usr/bin/env bash
  sudo nixos-firewall-tool open tcp 8081 19000 19001
  sudo nixos-firewall-tool open udp 8081 19000 19001

expo-start:
  #!/usr/bin/env bash
  cd app
  npx expo start
