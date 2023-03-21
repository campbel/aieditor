
import * as vscode from 'vscode';
const { Configuration, OpenAIApi } = require("openai");
import * as fs from 'fs';

// Get the home directory
const homeDir = require('os').homedir();

// Read openai api key from ~/.config/openai/token
const openaiApiKey = fs.readFileSync(`${homeDir}/.config/openai/token`, 'utf-8').trim();

// Create openai client
const openai = new OpenAIApi(new Configuration({
	apiKey: openaiApiKey,
}));

async function getCompletion(prompt: string): Promise<string> {
	try {
		const response = await openai.createCompletion({
			model: 'text-davinci-003',
			prompt: prompt,
			max_tokens: 4000-prompt.length,
			temperature: 0.7,
		});
		return response.data.choices[0].text;
	} catch (error: any) {
		if (error.response) {
		  console.log(error.response.status);
		  console.log(error.response.data);
		} else {
		  console.log(error.message);
		}
		throw error;
	}
}

function getPrompt(type: string, input: string, language: string, code: string) {
	switch (type) {
		case 'modify':
			return `Modify this ${language} code by:\n${input}\n\n${code}`;
		case 'generate':
			return `Write some ${language} code that:\n${input}`;
		default:
			throw new Error(`Unknown type: ${type}`);
	}
}

function handle(type: string) {

	let inputBox = ((type: string) => {
		switch (type) {
			case 'modify':
				return {
					placeHolder: "Fix the following code",
					prompt: "Prompt for the AI",
					value: ""
				};
			case 'generate':
				return {
					placeHolder: "Write a ruby program that does something",
					prompt: "Prompt for the AI",
					value: ""
				};
			default:
				throw new Error(`Unknown type: ${type}`);
			}
		})(type);

	return async () => {
		const editor = vscode.window.activeTextEditor;
		if (!editor) {
			vscode.window.showInformationMessage("Couldn't get active editor!");
			return;
		}
		const inputText = await vscode.window.showInputBox(inputBox);
		if (!inputText) {
			vscode.window.showInformationMessage("No prompt provided!");
			return;
		}
		const document = editor.document;
		const language = document.languageId;
		const selection = editor.selection;
		const code = document.getText(selection);
		console.log(`document: ${document} language: ${language} selection: ${selection} code: ${code}`);
		vscode.window.showInformationMessage('Fetching completion...');
		try {
			const completion = await getCompletion(`${getPrompt(type, inputText, language, code)}`);
			vscode.window.showInformationMessage('Got completion!');
			editor.edit(editBuilder => {
				if (type === 'modify') {
					editBuilder.replace(selection, completion);
				}
				if (type === 'generate') {
					editBuilder.insert(selection.end, completion);
				}
			});
		} catch (e) {
			vscode.window.showInformationMessage(`${e}`);
			console.log(e);
			return;
		}
	};
}

export async function activate(context: vscode.ExtensionContext) {
	context.subscriptions.push(vscode.commands.registerCommand('aieditor.modifyCode', handle('modify')));
	context.subscriptions.push(vscode.commands.registerCommand('aieditor.generateCode', handle('generate')));
}

// This method is called when your extension is deactivated
export function deactivate() {}
