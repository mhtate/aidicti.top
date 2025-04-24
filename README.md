# aidicti.top

# 📚 Telegram Dictionary Bot

A personal Telegram bot written in Go that helps you build and practice your own dictionary by translating words using the Oxford Dictionary and generating example sentences using ChatGPT.

## ✨ Features

- 🔍 Translate English words using Oxford Dictionary
- 🎙️ Send a **voice message** to translate a spoken word or phrase *(no Telegram Premium required)* — powered by Go speech-to-text
- 📖 Add translations to your private dictionary
- 🧠 Practice learned words by generating example sentences via ChatGPT
- 🗂️ Store your own word list for future reference

## 📦 Tech Stack

- **Go** (Golang)
- **[go-telegram-bot](https://github.com/go-telegram/bot)** for Telegram Bot API
- **Oxford Dictionaries API** for definitions and translations
- **OpenAI ChatGPT API** for sentence generation
- **Speech-to-Text** using Go libraries or external services (e.g., Whisper, Google STT)

## 🚀 Getting Started

### Prerequisites

- Go 1.20+
- Telegram Bot Token
- Oxford Dictionaries API credentials
- OpenAI API Key
- (Optional) Speech-to-text backend (e.g. Whisper CLI or Google Cloud STT)

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/your-username/telegram-dictionary-bot.git
   cd telegram-dictionary-bot
Set your environment variables:

bash
Copy
Edit
export TELEGRAM_BOT_TOKEN=your_token_here
export OXFORD_APP_ID=your_oxford_app_id
export OXFORD_APP_KEY=your_oxford_app_key
export OPENAI_API_KEY=your_openai_key
Run the bot:

bash
Copy
Edit
go run main.go

## 🛠 Usage
#/dict - Enables dictionary mode, where you can see word meanings.
#/pract – Ask ChatGPT to generate example sentences for words in your dictionary.

## 📸 Screenshots
🔍 Translating a word

🎙️ Sending a voice message

🧠 Practicing a word with ChatGPT

📖 Viewing your saved dictionary

## 📝 TODO
Store words in persistent database (currently in-memory/file)

Add spaced repetition algorithm for practice

Improve multi-language support

Web dashboard to view and edit dictionary

Offline Whisper support for speech-to-text

## 🤝 Contributing
Pull requests are welcome! For major changes, please open an issue first to discuss what you'd like to change.

## 📄 License
This project is licensed under the MIT License.
