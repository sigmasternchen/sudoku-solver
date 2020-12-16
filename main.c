#include <string.h>
#include <stdbool.h>
#include <stdlib.h>
#include <stdio.h>

#define MAX_LIST_LENGTH 9

typedef struct {
	int length;
	int list[MAX_LIST_LENGTH];
} list_t;

typedef struct {
	list_t field[9][9];
} field_t;

void setAll(list_t* list) {
	list->length = MAX_LIST_LENGTH;
	for(int i = 0; i < MAX_LIST_LENGTH; i++) {
		list->list[i] = i + 1;
	}
}

void output(field_t field) {
	for (int y = 0; y < 9; y++) {
		for (int x = 0; x < 9; x++) {
			if (field.field[x][y].length == 0) {
				putchar('#');
			} else if (field.field[x][y].length == 1) {
				putchar(field.field[x][y].list[0] + '0');
			} else {
				putchar('?');
			}
		}
		putchar('\n');
	}
}

field_t readField(const char* input) {
	int x = 0, y = 0;
	
	field_t field;
	
	size_t length = strlen(input);
	
	for (size_t i = 0; i < length; i++) {
		char c = input[i];
		bool inc = false;
		
		if (c == ' ') {
			setAll(&(field.field[x][y]));
			inc = true;
		} else if (c >= '1' && c <= '9') {
			field.field[x][y].length = 1;
			field.field[x][y].list[0] = c - '0';
			
			inc = true;
		}
		
		if (inc) {
			x++;
			if (x >= 9) {
				x = 0;
				y++;
			}
			if (y >= 9) {
				break;
			}
		}
	}
	
	return field;
}

static inline void removeFromList(list_t* list, int i) {
	for (int li = i + 1; li < list->length; li++) {
		list->list[li - 1] = list->list[li];
	}
	
	list->length--;
}

static inline int eliminate(field_t *field) {
	int eliminated = 0;
	
	// columns
	for (int x = 0; x < 9; x++) {
		unsigned long bitfield = 0x1ff;
		
		for (int y = 0; y < 9; y++) {
			list_t* list = &(field->field[x][y]);
			
			if (list->length == 1) {
				bitfield &= ~(1 << (list->list[0] - 1));
			}
		}
		
		for (int y = 0; y < 9; y++) {
			list_t* list = &(field->field[x][y]);
			
			if (list->length == 1) {
				continue;
			}
			
			for (int i = 0; i < list->length; i++) {
				if (!(bitfield & (1 << (list->list[i] - 1)))) {
					eliminated++;
					removeFromList(list, i);
					i--;
				}
			}
		}
	}
	
	// rows
	for (int y = 0; y < 9; y++) {
		unsigned long bitfield = 0x1ff;
		
		for (int x = 0; x < 9; x++) {
			list_t* list = &(field->field[x][y]);
			
			if (list->length == 1) {
				bitfield &= ~(1 << (list->list[0] - 1));
			}
		}
		
		for (int x = 0; x < 9; x++) {
			list_t* list = &(field->field[x][y]);
			
			if (list->length == 1) {
				continue;
			}
			
			for (int i = 0; i < list->length; i++) {
				if (!(bitfield & (1 << (list->list[i] - 1)))) {
					eliminated++;
					removeFromList(list, i);
					i--;
				}
			}
		}
	}
	
	for (int xs = 0; xs < 3; xs++) {
		for (int ys = 0; ys < 3; ys++) {
			unsigned long bitfield = 0x1ff;
			
			for (int x = xs * 3; x < (xs + 1) * 3; x++) {
				for (int y = ys * 3; y < (ys + 1) * 3; y++) {
					list_t* list = &(field->field[x][y]);
			
					if (list->length == 1) {
						bitfield &= ~(1 << (list->list[0] - 1));
					}
				}
			}
			
			for (int x = xs * 3; x < (xs + 1) * 3; x++) {
				for (int y = ys * 3; y < (ys + 1) * 3; y++) {
					list_t* list = &(field->field[x][y]);
			
					if (list->length == 1) {
						continue;
					}
					
					for (int i = 0; i < list->length; i++) {
						if (!(bitfield & (1 << (list->list[i] - 1)))) {
							eliminated++;
							removeFromList(list, i);
							i--;
						}
					}
				}
			}
		}
	}
	
	return eliminated;
}

typedef enum {
	solved, unsolved, error
} state_t;

typedef struct {
	state_t state;
	field_t field;
} solution_t;

static inline state_t getState(field_t field) {
	for (int x = 0; x < 9; x++) {
		for (int y = 0; y < 9; y++) {
			int l = field.field[x][y].length;
			switch(l) {
				case 0:
					return error;
				case 1:
					continue;
				default:
					return unsolved;
			}
		}
	}
	
	return solved;
}

typedef struct {
	int length;
	struct guess {
		int x;
		int y;
		int n;
	}* guesses;
} guesslist_t;

static inline guesslist_t possibleGuesses(field_t field) {
	guesslist_t list;
	list.guesses = NULL;
	list.length = 0;
	
	for (int x = 0; x < 9; x++) {
		for (int y = 0; y < 9; y++) {
			int l = field.field[x][y].length;
			
			if (l > 1) {
				void* tmp = realloc(list.guesses, (list.length + l) * sizeof(struct guess));
				if (tmp == NULL) {
					fprintf(stderr, "panic: malloc failed\n");
					exit(1);
				}
				list.guesses = tmp;
				for (int i = 0; i < l; i++) {
					list.guesses[list.length++] = (struct guess) {
						x, y, field.field[x][y].list[i]
					};
				}
			}
		}
	}
	
	return list;
}

static inline void freeGuesses(guesslist_t guesses) {
	free(guesses.guesses);
}

solution_t solve(field_t field) {
	int n;
	
	while ((n = eliminate(&field))) {
		printf("possibilities eliminated: %d\n", n);
	}
	
	state_t state = getState(field);
	
	if (state == solved || state == error) {
		return (solution_t) {
			state, field
		};
	} else {
		printf("no solution found using basic rules\n");
		
		guesslist_t list = possibleGuesses(field);
		
		for (int i = 0; i < list.length; i++) {
			field_t field_ = field;
			struct guess guess = list.guesses[i];
			
			printf("guessing: %d,%d = %d\n", guess.x + 1, guess.y + 1, guess.n);
			
			field_.field[guess.x][guess.y].length = 1;
			field_.field[guess.x][guess.y].list[0] = guess.n;
			
			solution_t solution = solve(field_);
			
			if (solution.state == solved) {
				freeGuesses(list);
				
				return solution;
			}
		}
		
		freeGuesses(list);
		
		return (solution_t) {
			error,
			field
		};
	}
}

const char* wikipedia = "\
53  7    \
6  195   \
 98    6 \
8   6   3\
4  8 3  1\
7   2   6\
 6    28 \
   419  5\
    8  79\
";

const char* medium = "\
47 53    \
 9     8 \
 1  2  5 \
1  7  5 4\
  39  1  \
   65 9 3\
95  172 6\
28      1\
7   64 9 \
";

const char* hard = "\
   29  4 \
  31 5 26\
  96     \
2    83  \
1    98 5\
 57      \
768     4\
    6 2 9\
    4   3\
";

const char* very_difficult = "\
 21 6 4  \
   5   9 \
4    2  1\
84  5    \
1       2\
    4  75\
7  6    4\
 3   9   \
  8 3 61 \
";

const char* hardest = "\
8        \
  36     \
 7  9 2  \
 5   7   \
    457  \
   1   3 \
  1    68\
  85   1 \
 9    4  \
";

int main() {
	field_t field = readField(hardest);
	output(field);
	
	solution_t solution = solve(field);
	
	printf("\n\n");
	
	switch(solution.state) {
		case error: 
			printf("error\n");
			break;
		case solved: 
			printf("solved\n");
			break;
		case unsolved: 
			printf("unsolved\n");
			break;
		default:
			break;
	}
	
	output(solution.field);
	
	return 0;
}
