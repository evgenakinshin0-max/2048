#!/usr/bin/env ruby
# 2048.rb
# encoding: UTF-8

require 'json'
require 'io/console'

COLORS = {
  reset: "\e[0m",
  bold: "\e[1m",
  red: "\e[91m",
  green: "\e[92m",
  yellow: "\e[93m",
  blue: "\e[94m",
  magenta: "\e[95m",
  cyan: "\e[96m",
  white: "\e[97m",
  bgRed: "\e[101m",
  bgGreen: "\e[102m",
  bgYellow: "\e[103m",
  bgBlue: "\e[104m",
  bgMagenta: "\e[105m",
  bgCyan: "\e[106m",
  bgWhite: "\e[107m"
}

def colorize(text, color)
  "#{COLORS[color]}#{text}#{COLORS[:reset]}"
end

TILE_COLORS = {
  2 => { bg: :bgYellow, fg: :black },
  4 => { bg: :bgBlue, fg: :white },
  8 => { bg: :bgCyan, fg: :black },
  16 => { bg: :bgGreen, fg: :white },
  32 => { bg: :bgMagenta, fg: :white },
  64 => { bg: :bgRed, fg: :white },
  128 => { bg: :bgYellow, fg: :black },
  256 => { bg: :bgBlue, fg: :white },
  512 => { bg: :bgCyan, fg: :black },
  1024 => { bg: :bgMagenta, fg: :white },
  2048 => { bg: :bgRed, fg: :white },
  4096 => { bg: :bgGreen, fg: :black },
  8192 => { bg: :bgBlue, fg: :white }
}

def get_tile_color(val)
  TILE_COLORS[val] || { bg: :bgWhite, fg: :black }
end

RECORD_FILE = File.join(Dir.home, '.2048_records.json')

def load_records
  return {} unless File.exist?(RECORD_FILE)
  JSON.parse(File.read(RECORD_FILE))
rescue
  {}
end

def save_records(records)
  File.write(RECORD_FILE, JSON.pretty_generate(records))
end

class Game2048
  attr_reader :size, :board, :score, :best, :history

  def initialize(size = 4)
    @size = size
    @board = Array.new(size) { Array.new(size, 0) }
    @score = 0
    @history = []
    @records = load_records
    @best = @records[size.to_s] || 0
    @rng = Random.new
  end

  def add_random_tile
    empty = []
    @size.times do |i|
      @size.times do |j|
        empty << [i, j] if @board[i][j] == 0
      end
    end
    return if empty.empty?
    i, j = empty.sample
    @board[i][j] = @rng.rand < 0.9 ? 2 : 4
  end

  def compress(row)
    res = row.reject { |x| x == 0 }
    res += [0] * (@size - res.length)
    res
  end

  def merge(row)
    res = []
    add = 0
    i = 0
    while i < row.length
      if i < row.length - 1 && row[i] == row[i + 1] && row[i] != 0
        res << row[i] * 2
        add += row[i] * 2
        i += 2
      else
        res << row[i]
        i += 1
      end
    end
    res += [0] * (@size - res.length)
    [res, add]
  end

  def rotate(board)
    board.transpose.map(&:reverse)
  end

  def move(dir)
    prev_board = @board.map(&:dup)
    prev_score = @score
    add = 0

    times = { 'w' => 0, 'd' => 1, 's' => 2, 'a' => 3 }[dir] || 0
    b = @board.map(&:dup)
    times.times { b = rotate(b) }

    b.size.times do |i|
      row = compress(b[i])
      merged, sc = merge(row)
      b[i] = compress(merged)
      add += sc
    end

    (4 - times) % 4.times { b = rotate(b) }

    if b != @board
      @history << { board: prev_board, score: prev_score }
      @board = b
      @score += add
      if @score > @best
        @best = @score
        @records[@size.to_s] = @best
        save_records(@records)
      end
      add_random_tile
      return true
    end
    false
  end

  def undo
    return false if @history.empty?
    last = @history.pop
    @board = last[:board]
    @score = last[:score]
    true
  end

  def is_win?
    @board.any? { |row| row.any? { |v| v >= 2048 } }
  end

  def is_game_over?
    @size.times do |i|
      @size.times do |j|
        return false if @board[i][j] == 0
        return false if i < @size - 1 && @board[i][j] == @board[i + 1][j]
        return false if j < @size - 1 && @board[i][j] == @board[i][j + 1]
      end
    end
    true
  end

  def display
    system('clear') || system('cls')
    puts colorize("🎮  2048  |  Размер #{@size}×#{@size}  |  Счёт: #{@score}  |  Лучший: #{@best}", :bold)
    puts '─' * (@size * 8 + 1)
    @size.times do |i|
      @size.times do |j|
        val = @board[i][j]
        if val == 0
          print colorize('      ', :bgWhite) + ' '
        else
          tc = get_tile_color(val)
          print colorize(val.to_s.rjust(6), tc[:bg]) + ' '
        end
        print '│' if j < @size - 1
      end
      puts
      puts '─' * (@size * 8 + 1) if i < @size - 1
    end
    puts '─' * (@size * 8 + 1)
    puts 'WASD — движение | U — отмена | Q — выход'
  end
end

def main
  size = 4
  i = 0
  while i < ARGV.length
    case ARGV[i]
    when '-s'
      size = ARGV[i+1].to_i
      i += 2
    when '-h', '--help'
      puts "Usage: ruby 2048.rb [options]\n  -s <N>   Size (4, 5, 6) default 4"
      return
    else
      i += 1
    end
  end
  size = 4 if size < 4 || size > 6

  game = Game2048.new(size)
  game.add_random_tile
  game.add_random_tile

  loop do
    game.display
    puts colorize("🎉 Поздравляем! Вы достигли 2048!", :bold) if game.is_win?
    puts colorize("💀 Игра окончена. Нажмите Q для выхода.", :red) if game.is_game_over?

    ch = STDIN.getch
    case ch
    when 'q', 'Q'
      puts colorize("Игра сохранена. До встречи!", :yellow)
      break
    when 'u', 'U'
      if game.undo
        puts colorize("Ход отменён.", :yellow)
      else
        puts colorize("Нечего отменять.", :yellow)
      end
    when 'w', 'W' then game.move('w')
    when 'a', 'A' then game.move('a')
    when 's', 'S' then game.move('s')
    when 'd', 'D' then game.move('d')
    end
  end
end

main if __FILE__ == $0
