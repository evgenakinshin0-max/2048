# 2048.py
#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import sys
import os
import random
import json
import copy
import argparse
from datetime import datetime
from pathlib import Path

# ANSI colors
COLORS = {
    'reset': '\033[0m',
    'bold': '\033[1m',
    'black': '\033[30m',
    'red': '\033[91m',
    'green': '\033[92m',
    'yellow': '\033[93m',
    'blue': '\033[94m',
    'magenta': '\033[95m',
    'cyan': '\033[96m',
    'white': '\033[97m',
    'bg_red': '\033[101m',
    'bg_green': '\033[102m',
    'bg_yellow': '\033[103m',
    'bg_blue': '\033[104m',
    'bg_magenta': '\033[105m',
    'bg_cyan': '\033[106m',
    'bg_white': '\033[107m',
}

def colorize(text, color):
    return f"{COLORS.get(color, '')}{text}{COLORS['reset']}"

# Цвета для плиток (фоновый + текст)
TILE_COLORS = {
    2: ('bg_yellow', 'black'),
    4: ('bg_blue', 'white'),
    8: ('bg_cyan', 'black'),
    16: ('bg_green', 'white'),
    32: ('bg_magenta', 'white'),
    64: ('bg_red', 'white'),
    128: ('bg_yellow', 'black'),
    256: ('bg_blue', 'white'),
    512: ('bg_cyan', 'black'),
    1024: ('bg_magenta', 'white'),
    2048: ('bg_red', 'white'),
    4096: ('bg_green', 'black'),
    8192: ('bg_blue', 'white'),
}

def get_tile_color(value):
    if value in TILE_COLORS:
        return TILE_COLORS[value]
    return ('bg_white', 'black')

RECORD_FILE = Path.home() / '.2048_records.json'
SAVE_FILE = Path.home() / '.2048_save.json'

def load_records():
    if RECORD_FILE.exists():
        with open(RECORD_FILE, 'r') as f:
            return json.load(f)
    return {}

def save_records(records):
    with open(RECORD_FILE, 'w') as f:
        json.dump(records, f, indent=2)

def load_save(size):
    if SAVE_FILE.exists():
        with open(SAVE_FILE, 'r') as f:
            data = json.load(f)
            if data.get('size') == size:
                return data
    return None

def save_game(state, size):
    data = {
        'size': size,
        'board': state['board'],
        'score': state['score'],
        'history': state['history'],
        'timestamp': datetime.now().isoformat()
    }
    with open(SAVE_FILE, 'w') as f:
        json.dump(data, f, indent=2)

class Game2048:
    def __init__(self, size=4):
        self.size = size
        self.board = [[0] * size for _ in range(size)]
        self.score = 0
        self.history = []  # для undo
        self.records = load_records()
        self.best = self.records.get(str(size), 0)

    def add_random_tile(self):
        empty = [(i, j) for i in range(self.size) for j in range(self.size) if self.board[i][j] == 0]
        if empty:
            i, j = random.choice(empty)
            self.board[i][j] = 2 if random.random() < 0.9 else 4

    def compress(self, row):
        new_row = [x for x in row if x != 0]
        return new_row + [0] * (self.size - len(new_row))

    def merge(self, row):
        score_add = 0
        new_row = []
        i = 0
        while i < len(row):
            if i < len(row) - 1 and row[i] == row[i + 1] and row[i] != 0:
                new_row.append(row[i] * 2)
                score_add += row[i] * 2
                i += 2
            else:
                new_row.append(row[i])
                i += 1
        new_row += [0] * (self.size - len(new_row))
        return new_row, score_add

    def move(self, direction):
        prev_board = copy.deepcopy(self.board)
        prev_score = self.score
        moved = False
        score_add = 0

        if direction == 'up':
            for j in range(self.size):
                col = [self.board[i][j] for i in range(self.size)]
                col = self.compress(col)
                col, add = self.merge(col)
                col = self.compress(col)
                for i in range(self.size):
                    self.board[i][j] = col[i]
                score_add += add
        elif direction == 'down':
            for j in range(self.size):
                col = [self.board[i][j] for i in range(self.size)]
                col = self.compress(col[::-1])[::-1]
                col, add = self.merge(col)
                col = self.compress(col[::-1])[::-1]
                for i in range(self.size):
                    self.board[i][j] = col[i]
                score_add += add
        elif direction == 'left':
            for i in range(self.size):
                row = self.compress(self.board[i])
                row, add = self.merge(row)
                row = self.compress(row)
                self.board[i] = row
                score_add += add
        elif direction == 'right':
            for i in range(self.size):
                row = self.compress(self.board[i][::-1])[::-1]
                row, add = self.merge(row)
                row = self.compress(row[::-1])[::-1]
                self.board[i] = row
                score_add += add

        if self.board != prev_board:
            self.history.append({'board': prev_board, 'score': prev_score})
            self.score += score_add
            if self.score > self.best:
                self.best = self.score
                key = str(self.size)
                self.records[key] = self.best
                save_records(self.records)
            self.add_random_tile()
            return True
        return False

    def undo(self):
        if not self.history:
            return False
        last = self.history.pop()
        self.board = last['board']
        self.score = last['score']
        return True

    def is_win(self):
        for i in range(self.size):
            for j in range(self.size):
                if self.board[i][j] >= 2048:
                    return True
        return False

    def is_game_over(self):
        for i in range(self.size):
            for j in range(self.size):
                if self.board[i][j] == 0:
                    return False
                if i < self.size - 1 and self.board[i][j] == self.board[i+1][j]:
                    return False
                if j < self.size - 1 and self.board[i][j] == self.board[i][j+1]:
                    return False
        return True

    def display(self):
        os.system('clear' if os.name == 'posix' else 'cls')
        print(colorize(f"🎮  2048  |  Размер {self.size}×{self.size}  |  Счёт: {self.score}  |  Лучший: {self.best}", 'bold'))
        print("─" * (self.size * 8 + 1))
        for i in range(self.size):
            row_display = []
            for j in range(self.size):
                val = self.board[i][j]
                if val == 0:
                    row_display.append(colorize("      ", 'bg_white') + ' ')
                else:
                    bg, fg = get_tile_color(val)
                    text = str(val).center(6)
                    row_display.append(colorize(text, fg) + ' ')
            print('│' + '│'.join(row_display) + '│')
            print("─" * (self.size * 8 + 1))
        print("WASD — движение | U — отмена | Q — выход")

    def save(self):
        save_game({'board': self.board, 'score': self.score, 'history': self.history}, self.size)

def main():
    parser = argparse.ArgumentParser(description="2048 – игра в терминале")
    parser.add_argument('-s', '--size', type=int, default=4, choices=[4, 5, 6], help='Размер поля (4, 5, 6)')
    args = parser.parse_args()

    size = args.size
    game = Game2048(size)

    # Попытка загрузить сохранение
    saved = load_save(size)
    if saved:
        game.board = saved['board']
        game.score = saved['score']
        game.history = saved.get('history', [])
    else:
        game.add_random_tile()
        game.add_random_tile()

    try:
        while True:
            game.display()
            if game.is_win():
                print(colorize("🎉 Поздравляем! Вы достигли 2048!", 'bold'))
            if game.is_game_over():
                print(colorize("💀 Игра окончена. Нажмите Q для выхода.", 'red'))
            key = sys.stdin.read(1).lower()
            if key == 'q':
                game.save()
                print(colorize("Игра сохранена. До встречи!", 'yellow'))
                break
            elif key == 'u':
                if game.undo():
                    print(colorize("Ход отменён.", 'yellow'))
                else:
                    print(colorize("Нечего отменять.", 'yellow'))
            elif key in ['w', 'a', 's', 'd']:
                game.move(key)
    except KeyboardInterrupt:
        game.save()
        print(colorize("\nИгра сохранена. До встречи!", 'yellow'))

if __name__ == '__main__':
    main()
