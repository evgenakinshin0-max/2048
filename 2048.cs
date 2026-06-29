// 2048.cs
using System;
using System.Collections.Generic;
using System.IO;
using System.Text.Json;
using System.Threading;

class Game2048
{
    static string Colorize(string text, string color)
    {
        string col = color switch
        {
            "bold" => "\x1b[1m",
            "red" => "\x1b[91m",
            "green" => "\x1b[92m",
            "yellow" => "\x1b[93m",
            "blue" => "\x1b[94m",
            "magenta" => "\x1b[95m",
            "cyan" => "\x1b[96m",
            "white" => "\x1b[97m",
            "bgRed" => "\x1b[101m",
            "bgGreen" => "\x1b[102m",
            "bgYellow" => "\x1b[103m",
            "bgBlue" => "\x1b[104m",
            "bgMagenta" => "\x1b[105m",
            "bgCyan" => "\x1b[106m",
            "bgWhite" => "\x1b[107m",
            _ => "\x1b[0m"
        };
        return col + text + "\x1b[0m";
    }

    static string ConfigFile => Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), ".2048_records.json");

    static Dictionary<string, int> LoadRecords()
    {
        if (!File.Exists(ConfigFile)) return new Dictionary<string, int>();
        string json = File.ReadAllText(ConfigFile);
        return JsonSerializer.Deserialize<Dictionary<string, int>>(json) ?? new Dictionary<string, int>();
    }

    static void SaveRecords(Dictionary<string, int> records)
    {
        string json = JsonSerializer.Serialize(records, new JsonSerializerOptions { WriteIndented = true });
        File.WriteAllText(ConfigFile, json);
    }

    class HistoryEntry
    {
        public int[,] Board { get; set; }
        public int Score { get; set; }
    }

    private int size;
    private int[,] board;
    private int score;
    private int best;
    private List<HistoryEntry> history;
    private Dictionary<string, int> records;
    private Random rng;

    public Game2048(int s = 4)
    {
        size = s;
        board = new int[size, size];
        score = 0;
        history = new List<HistoryEntry>();
        records = LoadRecords();
        best = records.GetValueOrDefault(size.ToString(), 0);
        rng = new Random();
    }

    void AddRandomTile()
    {
        var empty = new List<(int, int)>();
        for (int i = 0; i < size; i++)
            for (int j = 0; j < size; j++)
                if (board[i, j] == 0) empty.Add((i, j));
        if (empty.Count == 0) return;
        var (i, j) = empty[rng.Next(empty.Count)];
        board[i, j] = rng.NextDouble() < 0.9 ? 2 : 4;
    }

    int[] Compress(int[] row)
    {
        var res = new List<int>();
        foreach (var x in row) if (x != 0) res.Add(x);
        while (res.Count < size) res.Add(0);
        return res.ToArray();
    }

    (int[], int) Merge(int[] row)
    {
        var res = new List<int>();
        int add = 0;
        int i = 0;
        while (i < row.Length)
        {
            if (i < row.Length - 1 && row[i] == row[i + 1] && row[i] != 0)
            {
                res.Add(row[i] * 2);
                add += row[i] * 2;
                i += 2;
            }
            else
            {
                res.Add(row[i]);
                i++;
            }
        }
        while (res.Count < size) res.Add(0);
        return (res.ToArray(), add);
    }

    int[,] CopyBoard()
    {
        var b = new int[size, size];
        for (int i = 0; i < size; i++)
            for (int j = 0; j < size; j++)
                b[i, j] = board[i, j];
        return b;
    }

    bool EqualBoard(int[,] a, int[,] b)
    {
        for (int i = 0; i < size; i++)
            for (int j = 0; j < size; j++)
                if (a[i, j] != b[i, j]) return false;
        return true;
    }

    int[,] Rotate(int[,] b)
    {
        var res = new int[size, size];
        for (int i = 0; i < size; i++)
            for (int j = 0; j < size; j++)
                res[j, size - 1 - i] = b[i, j];
        return res;
    }

    public bool Move(string dir)
    {
        var prevBoard = CopyBoard();
        int prevScore = score;
        int addScore = 0;

        int times = dir switch
        {
            "w" => 0,
            "d" => 1,
            "s" => 2,
            "a" => 3,
            _ => 0
        };

        var b = CopyBoard();
        for (int t = 0; t < times; t++) b = Rotate(b);

        for (int i = 0; i < size; i++)
        {
            var row = new int[size];
            for (int j = 0; j < size; j++) row[j] = b[i, j];
            row = Compress(row);
            var (merged, add) = Merge(row);
            merged = Compress(merged);
            for (int j = 0; j < size; j++) b[i, j] = merged[j];
            addScore += add;
        }

        for (int t = 0; t < (4 - times) % 4; t++) b = Rotate(b);

        if (!EqualBoard(b, board))
        {
            history.Add(new HistoryEntry { Board = prevBoard, Score = prevScore });
            board = b;
            score += addScore;
            if (score > best)
            {
                best = score;
                records[size.ToString()] = best;
                SaveRecords(records);
            }
            AddRandomTile();
            return true;
        }
        return false;
    }

    public bool Undo()
    {
        if (history.Count == 0) return false;
        var last = history[history.Count - 1];
        history.RemoveAt(history.Count - 1);
        board = last.Board;
        score = last.Score;
        return true;
    }

    public bool IsWin()
    {
        for (int i = 0; i < size; i++)
            for (int j = 0; j < size; j++)
                if (board[i, j] >= 2048) return true;
        return false;
    }

    public bool IsGameOver()
    {
        for (int i = 0; i < size; i++)
            for (int j = 0; j < size; j++)
            {
                if (board[i, j] == 0) return false;
                if (i < size - 1 && board[i, j] == board[i + 1, j]) return false;
                if (j < size - 1 && board[i, j] == board[i, j + 1]) return false;
            }
        return true;
    }

    string GetTileColor(int val)
    {
        return val switch
        {
            2 => "bgYellow",
            4 => "bgBlue",
            8 => "bgCyan",
            16 => "bgGreen",
            32 => "bgMagenta",
            64 => "bgRed",
            128 => "bgYellow",
            256 => "bgBlue",
            512 => "bgCyan",
            1024 => "bgMagenta",
            2048 => "bgRed",
            4096 => "bgGreen",
            8192 => "bgBlue",
            _ => "bgWhite"
        };
    }

    public void Display()
    {
        Console.Clear();
        Console.WriteLine(Colorize($"🎮  2048  |  Размер {size}×{size}  |  Счёт: {score}  |  Лучший: {best}", "bold"));
        Console.WriteLine(new string('─', size * 8 + 1));
        for (int i = 0; i < size; i++)
        {
            for (int j = 0; j < size; j++)
            {
                int val = board[i, j];
                if (val == 0)
                    Console.Write(Colorize("      ", "bgWhite") + " ");
                else
                {
                    string bg = GetTileColor(val);
                    Console.Write(Colorize(val.ToString().PadLeft(6), bg) + " ");
                }
                if (j < size - 1) Console.Write("│");
            }
            Console.WriteLine();
            if (i < size - 1) Console.WriteLine(new string('─', size * 8 + 1));
        }
        Console.WriteLine(new string('─', size * 8 + 1));
        Console.WriteLine("WASD — движение | U — отмена | Q — выход");
    }

    static void Main(string[] args)
    {
        int size = 4;
        for (int i = 0; i < args.Length; i++)
        {
            if (args[i] == "-s" && i + 1 < args.Length)
                size = int.Parse(args[++i]);
            else if (args[i] == "-h" || args[i] == "--help")
            {
                Console.WriteLine("Usage: 2048 [options]\n  -s <N>   Size (4, 5, 6) default 4");
                return;
            }
        }
        if (size < 4 || size > 6) size = 4;

        var game = new Game2048(size);
        game.AddRandomTile();
        game.AddRandomTile();

        while (true)
        {
            game.Display();
            if (game.IsWin()) Console.WriteLine(Colorize("🎉 Поздравляем! Вы достигли 2048!", "bold"));
            if (game.IsGameOver()) Console.WriteLine(Colorize("💀 Игра окончена. Нажмите Q для выхода.", "red"));

            var key = Console.ReadKey(true).KeyChar;
            if (key == 'q' || key == 'Q') break;
            else if (key == 'u' || key == 'U')
            {
                if (game.Undo()) Console.WriteLine(Colorize("Ход отменён.", "yellow"));
                else Console.WriteLine(Colorize("Нечего отменять.", "yellow"));
            }
            else if (key == 'w' || key == 'W') game.Move("w");
            else if (key == 'a' || key == 'A') game.Move("a");
            else if (key == 's' || key == 'S') game.Move("s");
            else if (key == 'd' || key == 'D') game.Move("d");
        }
        Console.WriteLine(Colorize("Игра сохранена. До встречи!", "yellow"));
    }
}
