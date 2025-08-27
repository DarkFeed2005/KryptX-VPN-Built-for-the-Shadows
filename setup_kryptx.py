import os

# Define your folder and file structure
structure = {
    "cmd/client": ["main.go"],
    "cmd/server": ["main.go"],
    "internal/config": ["config.go", "vault.go"],
    "internal/network": ["wireguard.go", "dns.go", "killswitch.go"],
    "internal/gui": ["app.go", "theme.go", "components.go"],
    "internal/security": ["encryption.go", "auth.go"],
    "internal/utils": ["logger.go", "system.go"],
    "pkg/api": ["client.go", "server.go"],
    "web": ["index.html", "style.css", "script.js"],
    "configs": ["client.yaml", "server.yaml"],
    "scripts": ["build.sh", "install.sh"],
    "": ["go.mod", "go.sum", "Makefile", "README.md"]
}

def create_structure(base_path, structure):
    for folder, files in structure.items():
        folder_path = os.path.join(base_path, folder)
        if folder:  # Skip empty string for root files
            os.makedirs(folder_path, exist_ok=True)
        for file in files:
            file_path = os.path.join(base_path, folder, file)
            with open(file_path, 'w') as f:
                f.write("")  # Create empty file

if __name__ == "__main__":
    root = os.getcwd()  # Current directory
    create_structure(root, structure)
    print("[✔] KryptX folder structure created successfully.")