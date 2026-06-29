// 2048.js
#!/usr/bin/env node
'use strict';

const fs = require('fs');
const path = require('path');
const os = require('os');
const readline = require('readline');

const COLORS = {
    reset: '\x1b[0m',
    bold: '\x1b[1m',
    red: '\x1b[91m',
    green: '\x1b[92m',
    yellow: '\x1b[93m',
    blue: '\x1b[94m',
    magenta: '\x1b[95m',
    cyan: '\x1b[96m',
    white: '\x1b[97m',
    bgRed: '\x1b[101m',
    bgGreen: '\x1b[102m',
    bgYellow: '\x1b[103m',
    bgBlue: '\x1b[104m',
    bgMagenta: '\x1b[105m',
    bgCyan: '\x1b[106m',
    bgWhite: '\x1b[107m'
};

function colorize(text, color) {
    return (COLORS[color] || '') + text + COLORS.reset;
}

const TILE_COLORS = {
    2: { bg: 'bgYellow', fg: 'black' },
    4: { bg: 'bgBlue', fg: 'white' },
    8: { bg: 'bgCyan', fg: 'black' },
    16: { bg: 'bgGreen', fg: 'white' },
    32: { bg: 'bgMagenta', fg: 'white' },
    64: { bg: 'bgRed', fg: 'white' },
    128: { bg: 'bgYellow', fg: 'black' },
    256: { bg: 'bgBlue', fg: 'white' },
    512: { bg: 'bgCyan', fg: 'black' },
    1024: { bg: 'bgMagenta', fg: 'white' },
    2048: { bg: 'bgRed', fg: 'white' },
    4096: { bg: 'bgGreen', fg: 'black' },
    8192: { bg: 'bgBlue', fg: 'white' },
};

function getTileColor(val) {
    return TILE_COLORS[val] || { bg: 'bgWhite', fg: 'black' };
}

const RECORD_FILE = path.join(os.homedir(), '.2048_records.json');

function loadRecords() {
    try {
        return JSON.parse(fs.readFileSync(RECORD_FILE, 'utf8'));
    } catch { return {}; }
}

function saveRecords(records) {
    fs.writeFileSync(RECORD_FILE, JSON.stringify(records, null, 2));
}

class Game2048 {
    constructor(size = 4) {
        this.size = size;
        this.board = Array.from({ length: size }, () => Array(size).fill(0));
        this.score = 0;
        this.history = [];
        this.records = loadRecords();
        this.best = this.records[String(size)] || 0;
    }

    addRandomTile() {
        const empty = [];
        for (let i = 0; i < this.size; i++) {
            for (let j = 0; j < this.size; j++) {
                if (this.board[i][j] === 0) empty.push([i, j]);
            }
        }
        if (empty.length === 0) return;
        const [i, j] = empty[Math.floor(Math.random() * empty.length)];
        this.board[i][j] = Math.random() < 0.9 ? 2 : 4;
    }

    compress(row) {
        const res = row.filter(x => x !== 0);
        while (res.length < this.size) res.push(0);
        return res;
    }

    merge(row) {
        const res = [];
        let add = 0;
        let i = 0;
        while (i < row.length) {
            if (i < row.length - 1 && row[i] === row[i + 1] && row[i] !== 0) {
                res.push(row[i] * 2);
                add += row[i] * 2;
                i += 2;
            } else {
                res.push(row[i]);
                i++;
            }
        }
        while (res.length < this.size) res.push(0);
        return [res, add];
    }

    move(dir) {
        const prevBoard = this.board.map(row => [...row]);
        const prevScore = this.score;
        let add = 0;

        const rotate = (board) => {
            const n = board.length;
            const result = Array.from({ length: n }, () => Array(n).fill(0));
            for (let i = 0; i < n; i++) {
                for (let j = 0; j < n; j++) {
                    result[j][n - 1 - i] = board[i][j];
                }
            }
            return result;
        };

        let board = this.board.map(row => [...row]);
        let times = 0;
        if (dir === 'w') times = 0;
        else if (dir === 'd') times = 1;
        else if (dir === 's') times = 2;
        else if (dir === 'a') times = 3;

        for (let t = 0; t < times; t++) board = rotate(board);

        for (let i = 0; i < this.size; i++) {
            const row = this.compress(board[i]);
            const [merged, scoreAdd] = this.merge(row);
            const finalRow = this.compress(merged);
            board[i] = finalRow;
            add += scoreAdd;
        }

        for (let t = 0; t < (4 - times) % 4; t++) board = rotate(board);

        if (board.some((row, i) => row.some((val, j) => val !== this.board[i][j]))) {
            this.history.push({ board: prevBoard.map(row => [...row]), score: prevScore });
            this.board = board;
            this.score += add;
            if (this.score > this.best) {
                this.best = this.score;
                this.records[String(this.size)] = this.best;
                saveRecords(this.records);
            }
            this.addRandomTile();
            return true;
        }
        return false;
    }

    undo() {
        if (this.history.length === 0) return false;
        const last = this.history.pop();
        this.board = last.board;
        this.score = last.score;
        return true;
    }

    isWin() {
        for (const row of this.board) {
            for (const val of row) {
                if (val >= 2048) return true;
            }
        }
        return false;
    }

    isGameOver() {
        for (let i = 0; i < this.size; i++) {
            for (let j = 0; j < this.size; j++) {
                if (this.board[i][j] === 0) return false;
                if (i < this.size - 1 && this.board[i][j] === this.board[i + 1][j]) return false;
                if (j < this.size - 1 && this.board[i][j] === this.board[i][j + 1]) return false;
            }
        }
        return true;
    }

    display() {
        console.clear();
        console.log(colorize(`🎮  2048  |  Размер ${this.size}×${this.size}  |  Счёт: ${this.score}  |  Лучший: ${this.best}`, 'bold'));
        console.log('─'.repeat(this.size * 8 + 1));
        for (let i = 0; i < this.size; i++) {
            for (let j = 0; j < this.size; j++) {
                const val = this.board[i][j];
                if (val === 0) {
                    process.stdout.write(colorize('      ', 'bgWhite') + ' ');
                } else {
                    const tc = getTileColor(val);
                    const text = String(val).padStart(6);
                    process.stdout.write(colorize(text, tc.bg) + ' ');
                }
                if (j < this.size - 1) process.stdout.write('│');
            }
            console.log();
            if (i < this.size - 1) console.log('─'.repeat(this.size * 8 + 1));
        }
        console.log('─'.repeat(this.size * 8 + 1));
        console.log('WASD — движение | U — отмена | Q — выход');
    }
}

async function main() {
    const args = process.argv.slice(2);
    let size = 4;
    for (let i = 0; i < args.length; i++) {
        if (args[i] === '-s' && i + 1 < args.length) {
            size = parseInt(args[++i]);
        } else if (args[i] === '-h' || args[i] === '--help') {
            console.log('Usage: node 2048.js [options]\n  -s <N>   Size (4, 5, 6) default 4');
            return;
        }
    }
    if (size < 4 || size > 6) size = 4;

    const game = new Game2048(size);
    game.addRandomTile();
    game.addRandomTile();

    const rl = readline.createInterface({
        input: process.stdin,
        output: process.stdout,
        terminal: true
    });

    readline.emitKeypressEvents(process.stdin);
    process.stdin.setRawMode(true);

    process.stdin.on('keypress', (str, key) => {
        if (key && key.name === 'q') {
            console.log(colorize('Игра сохранена. До встречи!', 'yellow'));
            process.exit(0);
        } else if (key && key.name === 'u') {
            if (game.undo()) console.log(colorize('Ход отменён.', 'yellow'));
            else console.log(colorize('Нечего отменять.', 'yellow'));
        } else if (str === 'w' || key && key.name === 'up') {
            game.move('w');
        } else if (str === 'a' || key && key.name === 'left') {
            game.move('a');
        } else if (str === 's' || key && key.name === 'down') {
            game.move('s');
        } else if (str === 'd' || key && key.name === 'right') {
            game.move('d');
        }
        game.display();
        if (game.isWin()) console.log(colorize('🎉 Поздравляем! Вы достигли 2048!', 'bold'));
        if (game.isGameOver()) console.log(colorize('💀 Игра окончена. Нажмите Q для выхода.', 'red'));
    });

    game.display();
}

main().catch(err => console.error(err));
