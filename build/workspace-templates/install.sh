#!/bin/bash

# 下载 code-server
mkdir code-server && cd code-server
wget https://github.com/coder/code-server/releases/download/v4.9.0/code-server-4.9.0-linux-amd64.tar.gz
cd -

# 下载字体文件
mkdir fonts && cd fonts
wget https://github.com/adobe-fonts/source-code-pro/releases/download/variable-fonts/SourceCodeVariable-Italic.otf
wget https://github.com/adobe-fonts/source-code-pro/releases/download/variable-fonts/SourceCodeVariable-Roman.otf
cd -