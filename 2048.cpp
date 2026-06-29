// 2048.cpp
#include <iostream>
#include <vector>
#include <string>
#include <random>
#include <algorithm>
#include <thread>
#include <chrono>
#include <fstream>
#include <json/json.h> // sudo apt-get install libjsoncpp-dev
#include <termios.h>
#include <unistd.h>

using namespace std;

const string RESET = "\033[0m";
const string BOLD = "\033[1m";
const string RED = "\033[91m";
const string GREEN = "\033[92m";
const string YELLOW = "\033[93m";
const string BLUE = "\033[94m";
const string MAGENTA = "\033[95m";
const string CYAN = "\033[96m";
const string WHITE = "\033[97m";
const string BG_RED = "\033[101m";
const string BG_GREEN = "\033[102m";
const string BG_YELLOW = "\033[103m";
const string BG_BLUE = "\033[104m";
const string BG_MAGENTA = "\033[105m";
const string BG_CYAN = "\033[106m";
const string BG_WHITE = "\033[107m";

string colorize(const string& text, const string& color) {
    return color + text + RESET;
}

struct TileColors {
    string bg;
    string fg;
};

TileColors getTileColor(int value) {
    switch (value) {
        case 2: return {BG_YELLOW, "black"};
        case 4: return {BG_BLUE, WHITE};
        case 8: return {BG_CYAN, "black"};
        case 16: return {BG_GREEN, WHITE};
        case 32: return {BG_MAGENTA, WHITE};
        case 64: return {BG_RED, WHITE};
        case 128: return {BG_YELLOW, "black"};
        case 256: return {BG_BLUE, WHITE};
        case 512: return {BG_CYAN, "black"};
        case 1024: return {BG_MAGENTA, WHITE};
        case 2048: return {BG_RED, WHITE};
        case 4096: return {BG_GREEN, "black"};
        case 8192: return {BG_BLUE, WHITE};
        default: return {BG_WHITE, "black"};
    }
}

string getHomeDir() {
    const char* home = getenv("HOME");
    if (!home) home = getenv("USERPROFILE");
    return string(home);
}

string getRecordFile() {
    return getHomeDir() + "/.2048_records.json";
}

Json::Value loadRecords() {
    ifstream f(getRecordFile());
    Json::Value root;
    if (!f) return root;
    f >> root;
    return root;
}

void saveRecords(const Json::Value& records) {
    ofstream f(getRecordFile());
    f << records.toStyledString();
}

class Game2048 {
public:
    int size;
    vector<vector<int>> board;
    int score;
    int best;
    vector<pair<vector<vector<int>>, int>> history;
    Json::Value records;
    mt19937 rng;

    Game2048(int s = 4) : size(s), score(0), best(0), rng(random_device{}()) {
        board = vector<vector<int>>(size, vector<int>(size, 0));
        records = loadRecords();
        best = records[to_string(size)].asInt();
    }

    void addRandomTile() {
        vector<pair<int,int>> empty;
        for (int i = 0; i < size; ++i)
            for (int j = 0; j < size; ++j)
                if (board[i][j] == 0) empty.push_back({i, j});
        if (empty.empty()) return;
        auto [i, j] = empty[uniform_int_distribution<>(0, empty.size()-1)(rng)];
        board[i][j] = (uniform_real_distribution<>(0, 1)(rng) < 0.9) ? 2 : 4;
    }

    vector<int> compress(const vector<int>& row) {
        vector<int> res;
        for (int x : row) if (x != 0) res.push_back(x);
        while (res.size() < size) res.push_back(0);
        return res;
    }

    pair<vector<int>, int> merge(const vector<int>& row) {
        vector<int> res;
        int add = 0;
        int i = 0;
        while (i < row.size()) {
            if (i < row.size()-1 && row[i] == row[i+1] && row[i] != 0) {
                res.push_back(row[i]*2);
                add += row[i]*2;
                i += 2;
            } else {
                res.push_back(row[i]);
                i++;
            }
        }
        while (res.size() < size) res.push_back(0);
        return {res, add};
    }

    bool move(const string& dir) {
        auto prevBoard = board;
        int prevScore = score;
        int add = 0;

        if (dir == "w") {
            for (int j = 0; j < size; ++j) {
                vector<int> col;
                for (int i = 0; i < size; ++i) col.push_back(board[i][j]);
                col = compress(col);
                auto [merged, a] = merge(col);
                merged = compress(merged);
                for (int i = 0; i < size; ++i) board[i][j] = merged[i];
                add += a;
            }
        } else if (dir == "s") {
            for (int j = 0; j < size; ++j) {
                vector<int> col;
                for (int i = size-1; i >= 0; --i) col.push_back(board[i][j]);
                col = compress(col);
                auto [merged, a] = merge(col);
                merged = compress(merged);
                for (int i = 0; i < size; ++i) board[size-1-i][j] = merged[i];
                add += a;
            }
        } else if (dir == "a") {
            for (int i = 0; i < size; ++i) {
                auto row = compress(board[i]);
                auto [merged, a] = merge(row);
                merged = compress(merged);
                board[i] = merged;
                add += a;
            }
        } else if (dir == "d") {
            for (int i = 0; i < size; ++i) {
                auto row = board[i];
                reverse(row.begin(), row.end());
                row = compress(row);
                auto [merged, a] = merge(row);
                merged = compress(merged);
                reverse(merged.begin(), merged.end());
                board[i] = merged;
                add += a;
            }
        }

        if (board != prevBoard) {
            history.push_back({prevBoard, prevScore});
            score += add;
            if (score > best) {
                best = score;
                records[to_string(size)] = best;
                saveRecords(records);
            }
            addRandomTile();
            return true;
        }
        return false;
    }

    bool undo() {
        if (history.empty()) return false;
        auto [b, s] = history.back();
        history.pop_back();
        board = b;
        score = s;
        return true;
    }

    bool isWin() {
        for (auto& row : board)
            for (int v : row)
                if (v >= 2048) return true;
        return false;
    }

    bool isGameOver() {
        for (int i = 0; i < size; ++i)
            for (int j = 0; j < size; ++j) {
                if (board[i][j] == 0) return false;
                if (i < size-1 && board[i][j] == board[i+1][j]) return false;
                if (j < size-1 && board[i][j] == board[i][j+1]) return false;
            }
        return true;
    }

    void display() {
        cout << "\033[2J\033[1;1H";
        cout << colorize("🎮  2048  |  Размер " + to_string(size) + "×" + to_string(size) +
                         "  |  Счёт: " + to_string(score) + "  |  Лучший: " + to_string(best), BOLD) << endl;
        cout << string(size * 8 + 1, '─') << endl;
        for (int i = 0; i < size; ++i) {
            for (int j = 0; j < size; ++j) {
                int val = board[i][j];
                if (val == 0) {
                    cout << colorize("      ", BG_WHITE) << ' ';
                } else {
                    auto [bg, fg] = getTileColor(val);
                    string text = to_string(val);
                    text = string(6 - text.length(), ' ') + text;
                    cout << colorize(text, bg) << ' ';
                }
                if (j < size-1) cout << '│';
            }
            cout << endl;
            if (i < size-1) cout << string(size * 8 + 1, '─') << endl;
        }
        cout << string(size * 8 + 1, '─') << endl;
        cout << "WASD — движение | U — отмена | Q — выход" << endl;
    }
};

char getch() {
    struct termios oldt, newt;
    char ch;
    tcgetattr(STDIN_FILENO, &oldt);
    newt = oldt;
    newt.c_lflag &= ~(ICANON | ECHO);
    tcsetattr(STDIN_FILENO, TCSANOW, &newt);
    ch = getchar();
    tcsetattr(STDIN_FILENO, TCSANOW, &oldt);
    return ch;
}

int main(int argc, char* argv[]) {
    int size = 4;
    for (int i = 1; i < argc; ++i) {
        string arg = argv[i];
        if ((arg == "-s" || arg == "--size") && i+1 < argc) {
            size = stoi(argv[++i]);
        } else if (arg == "-h" || arg == "--help") {
            cout << "Usage: 2048 [options]\n  -s <N>   Size (4, 5, 6) default 4\n";
            return 0;
        }
    }
    if (size < 4 || size > 6) size = 4;

    Game2048 game(size);
    game.addRandomTile();
    game.addRandomTile();

    while (true) {
        game.display();
        if (game.isWin()) cout << colorize("🎉 Поздравляем! Вы достигли 2048!", BOLD) << endl;
        if (game.isGameOver()) cout << colorize("💀 Игра окончена. Нажмите Q для выхода.", RED) << endl;

        char ch = getch();
        if (ch == 'q' || ch == 'Q') {
            cout << colorize("Игра сохранена. До встречи!", YELLOW) << endl;
            break;
        } else if (ch == 'u' || ch == 'U') {
            if (game.undo()) cout << colorize("Ход отменён.", YELLOW) << endl;
            else cout << colorize("Нечего отменять.", YELLOW) << endl;
        } else if (ch == 'w' || ch == 'W') game.move("w");
        else if (ch == 'a' || ch == 'A') game.move("a");
        else if (ch == 's' || ch == 'S') game.move("s");
        else if (ch == 'd' || ch == 'D') game.move("d");
    }
    return 0;
}
