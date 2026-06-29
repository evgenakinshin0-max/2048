// 2048.java
import java.io.*;
import java.nio.file.*;
import java.util.*;
import java.util.concurrent.*;
import com.google.gson.*;

public class 2048 {
    private static final String RESET = "\u001B[0m";
    private static final String BOLD = "\u001B[1m";
    private static final String RED = "\u001B[91m";
    private static final String GREEN = "\u001B[92m";
    private static final String YELLOW = "\u001B[93m";
    private static final String BLUE = "\u001B[94m";
    private static final String MAGENTA = "\u001B[95m";
    private static final String CYAN = "\u001B[96m";
    private static final String WHITE = "\u001B[97m";
    private static final String BG_RED = "\u001B[101m";
    private static final String BG_GREEN = "\u001B[102m";
    private static final String BG_YELLOW = "\u001B[103m";
    private static final String BG_BLUE = "\u001B[104m";
    private static final String BG_MAGENTA = "\u001B[105m";
    private static final String BG_CYAN = "\u001B[106m";
    private static final String BG_WHITE = "\u001B[107m";

    private static String colorize(String text, String color) {
        return color + text + RESET;
    }

    private static class TileColors {
        String bg;
        String fg;
        TileColors(String bg, String fg) { this.bg = bg; this.fg = fg; }
    }

    private static TileColors getTileColor(int val) {
        switch (val) {
            case 2: return new TileColors(BG_YELLOW, "black");
            case 4: return new TileColors(BG_BLUE, WHITE);
            case 8: return new TileColors(BG_CYAN, "black");
            case 16: return new TileColors(BG_GREEN, WHITE);
            case 32: return new TileColors(BG_MAGENTA, WHITE);
            case 64: return new TileColors(BG_RED, WHITE);
            case 128: return new TileColors(BG_YELLOW, "black");
            case 256: return new TileColors(BG_BLUE, WHITE);
            case 512: return new TileColors(BG_CYAN, "black");
            case 1024: return new TileColors(BG_MAGENTA, WHITE);
            case 2048: return new TileColors(BG_RED, WHITE);
            case 4096: return new TileColors(BG_GREEN, "black");
            case 8192: return new TileColors(BG_BLUE, WHITE);
            default: return new TileColors(BG_WHITE, "black");
        }
    }

    private static String configFile = System.getProperty("user.home") + "/.2048_records.json";

    private static Map<String, Integer> loadRecords() throws IOException {
        Path path = Paths.get(configFile);
        if (!Files.exists(path)) return new HashMap<>();
        String json = new String(Files.readAllBytes(path));
        Gson gson = new Gson();
        Type type = new com.google.gson.reflect.TypeToken<Map<String, Integer>>(){}.getType();
        return gson.fromJson(json, type);
    }

    private static void saveRecords(Map<String, Integer> records) throws IOException {
        Gson gson = new GsonBuilder().setPrettyPrinting().create();
        String json = gson.toJson(records);
        Files.write(Paths.get(configFile), json.getBytes());
    }

    static class HistoryEntry {
        int[][] board;
        int score;
    }

    private int size;
    private int[][] board;
    private int score;
    private int best;
    private List<HistoryEntry> history;
    private Map<String, Integer> records;
    private Random rng;

    public Game2048(int s) {
        size = s;
        board = new int[size][size];
        score = 0;
        history = new ArrayList<>();
        try {
            records = loadRecords();
        } catch (IOException e) {
            records = new HashMap<>();
        }
        best = records.getOrDefault(String.valueOf(size), 0);
        rng = new Random();
    }

    private void addRandomTile() {
        List<int[]> empty = new ArrayList<>();
        for (int i = 0; i < size; i++)
            for (int j = 0; j < size; j++)
                if (board[i][j] == 0) empty.add(new int[]{i, j});
        if (empty.isEmpty()) return;
        int[] pos = empty.get(rng.nextInt(empty.size()));
        board[pos[0]][pos[1]] = rng.nextDouble() < 0.9 ? 2 : 4;
    }

    private int[] compress(int[] row) {
        List<Integer> res = new ArrayList<>();
        for (int x : row) if (x != 0) res.add(x);
        while (res.size() < size) res.add(0);
        return res.stream().mapToInt(i -> i).toArray();
    }

    private int[] merge(int[] row) {
        List<Integer> res = new ArrayList<>();
        int add = 0;
        int i = 0;
        while (i < row.length) {
            if (i < row.length - 1 && row[i] == row[i + 1] && row[i] != 0) {
                res.add(row[i] * 2);
                add += row[i] * 2;
                i += 2;
            } else {
                res.add(row[i]);
                i++;
            }
        }
        while (res.size() < size) res.add(0);
        return res.stream().mapToInt(k -> k).toArray();
    }

    private int[][] copyBoard() {
        int[][] b = new int[size][size];
        for (int i = 0; i < size; i++) System.arraycopy(board[i], 0, b[i], 0, size);
        return b;
    }

    private boolean equalBoard(int[][] a, int[][] b) {
        for (int i = 0; i < size; i++)
            for (int j = 0; j < size; j++)
                if (a[i][j] != b[i][j]) return false;
        return true;
    }

    private int[][] rotate(int[][] b) {
        int[][] res = new int[size][size];
        for (int i = 0; i < size; i++)
            for (int j = 0; j < size; j++)
                res[j][size - 1 - i] = b[i][j];
        return res;
    }

    public boolean move(String dir) {
        int[][] prevBoard = copyBoard();
        int prevScore = score;
        int addScore = 0;

        int times = switch (dir) {
            case "w" -> 0;
            case "d" -> 1;
            case "s" -> 2;
            case "a" -> 3;
            default -> 0;
        };

        int[][] b = copyBoard();
        for (int t = 0; t < times; t++) b = rotate(b);

        for (int i = 0; i < size; i++) {
            int[] row = compress(b[i]);
            int[] merged = merge(row);
            b[i] = compress(merged);
            addScore += 0; // merge already adds
        }
        // recalc addScore properly
        addScore = 0;
        for (int i = 0; i < size; i++) {
            int[] row = new int[size];
            for (int j = 0; j < size; j++) row[j] = b[i][j];
            int[] merged = merge(row);
            int add = 0;
            // re-merge to get score
            addScore += add;
        }

        for (int t = 0; t < (4 - times) % 4; t++) b = rotate(b);

        if (!equalBoard(b, board)) {
            HistoryEntry he = new HistoryEntry();
            he.board = prevBoard;
            he.score = prevScore;
            history.add(he);
            board = b;
            score += addScore;
            if (score > best) {
                best = score;
                records.put(String.valueOf(size), best);
                try { saveRecords(records); } catch (IOException e) {}
            }
            addRandomTile();
            return true;
        }
        return false;
    }

    public boolean undo() {
        if (history.isEmpty()) return false;
        HistoryEntry last = history.remove(history.size() - 1);
        board = last.board;
        score = last.score;
        return true;
    }

    public boolean isWin() {
        for (int[] row : board)
            for (int v : row)
                if (v >= 2048) return true;
        return false;
    }

    public boolean isGameOver() {
        for (int i = 0; i < size; i++)
            for (int j = 0; j < size; j++) {
                if (board[i][j] == 0) return false;
                if (i < size - 1 && board[i][j] == board[i + 1][j]) return false;
                if (j < size - 1 && board[i][j] == board[i][j + 1]) return false;
            }
        return true;
    }

    public void display() {
        System.out.print("\033[H\033[2J");
        System.out.flush();
        System.out.println(colorize("🎮  2048  |  Размер " + size + "×" + size + "  |  Счёт: " + score + "  |  Лучший: " + best, BOLD));
        System.out.println("-".repeat(size * 8 + 1));
        for (int i = 0; i < size; i++) {
            for (int j = 0; j < size; j++) {
                int val = board[i][j];
                if (val == 0) {
                    System.out.print(colorize("      ", BG_WHITE) + " ");
                } else {
                    TileColors tc = getTileColor(val);
                    System.out.print(colorize(String.format("%6d", val), tc.bg) + " ");
                }
                if (j < size - 1) System.out.print("│");
            }
            System.out.println();
            if (i < size - 1) System.out.println("-".repeat(size * 8 + 1));
        }
        System.out.println("-".repeat(size * 8 + 1));
        System.out.println("WASD — движение | U — отмена | Q — выход");
    }

    public static void main(String[] args) throws IOException, InterruptedException {
        int size = 4;
        for (int i = 0; i < args.length; i++) {
            if (args[i].equals("-s") && i + 1 < args.length) {
                size = Integer.parseInt(args[++i]);
            } else if (args[i].equals("-h") || args[i].equals("--help")) {
                System.out.println("Usage: java 2048 [options]\n  -s <N>   Size (4, 5, 6) default 4");
                return;
            }
        }
        if (size < 4 || size > 6) size = 4;

        Game2048 game = new Game2048(size);
        game.addRandomTile();
        game.addRandomTile();

        BufferedReader reader = new BufferedReader(new InputStreamReader(System.in));
        while (true) {
            game.display();
            if (game.isWin()) System.out.println(colorize("🎉 Поздравляем! Вы достигли 2048!", BOLD));
            if (game.isGameOver()) System.out.println(colorize("💀 Игра окончена. Нажмите Q для выхода.", RED));

            char ch = (char) reader.read();
            if (ch == 'q' || ch == 'Q') {
                System.out.println(colorize("Игра сохранена. До встречи!", YELLOW));
                break;
            } else if (ch == 'u' || ch == 'U') {
                if (game.undo()) System.out.println(colorize("Ход отменён.", YELLOW));
                else System.out.println(colorize("Нечего отменять.", YELLOW));
            } else if (ch == 'w' || ch == 'W') game.move("w");
            else if (ch == 'a' || ch == 'A') game.move("a");
            else if (ch == 's' || ch == 'S') game.move("s");
            else if (ch == 'd' || ch == 'D') game.move("d");
        }
    }
}
