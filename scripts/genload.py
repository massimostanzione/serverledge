import numpy as np
import random
import threading
import time
import os

IP = "192.168.122.21"
PORT = 1324

# Impostare il seed per generare la stessa sequenza ogni volta
seed_value = 12345
np.random.seed(seed_value)
random.seed(seed_value)

# Definizione mappa function_to_list
func_to_list = {"f1":[], "f2":[], "f3":[], "f4":[], "f5":[]}

# Definizione delle classi
classes = [
    {"name": "standard", "arrival_weight": 0.7},
    {"name": "critical-1", "arrival_weight": 0.1},
    {"name": "critical-2", "arrival_weight": 0.1},
    {"name": "batch", "arrival_weight": 0.1}
]

# Definizione delle funzioni e dei tassi di arrivo
functions = [
    {"name": "f1", "rate": 5, "param": 8100},
    {"name": "f2", "rate": 13, "param": 6000},
    {"name": "f3", "rate": 39, "param": 7200},
    {"name": "f4", "rate": 3, "param": 6550},
    {"name": "f5", "rate": 11, "param": 8500}
]

# Calcolo della distribuzione cumulativa per la selezione ponderata delle classi
weights = [cls['arrival_weight'] for cls in classes]
cumulative_weights = np.cumsum(weights)

def generate_poisson_arrivals(rate, duration):
    return np.cumsum(np.random.exponential(1/rate, int(rate * duration)))

def select_class():
    rand_value = random.uniform(0, cumulative_weights[-1])
    for i, weight in enumerate(cumulative_weights):
        if rand_value <= weight:
            return classes[i]['name']
    return classes[-1]['name']  # Default to last class if nothing matched

class ArrivalGenerator(threading.Thread):
    def __init__(self, function, duration):
        super().__init__()
        self.function = function
        self.duration = duration

    def run(self):
        arrivals = generate_poisson_arrivals(self.function['rate'], self.duration)
        for arrival_time in arrivals:
            class_name = select_class()
            #print(f"Function: {self.function['name']}, Arrival Time: {arrival_time:.4f}, Class: {class_name}")
            func_to_list[self.function['name']].append((arrival_time, self.function['name'], self.function['param'], class_name))

# Durata per la simulazione (ad esempio 10 secondi)
duration = 0.1

# Creazione e avvio dei thread per ogni funzione
threads = []
for func in functions:
    thread = ArrivalGenerator(func, duration)
    threads.append(thread)
    thread.start()

# Attendere la conclusione di tutti i thread
for thread in threads:
    thread.join()

complete_list = []
for key in func_to_list.keys():
    for t in func_to_list[key]:
        complete_list.append(t)

# Ordinamento della lista in-place
complete_list.sort(key=lambda x: x[0])

# Funzione che verrà eseguita dal thread per stampare il nome della funzione
def invoke_function(function_name, param, class_name):
    command = f"bin/serverledge-cli invoke -H {IP} -P {PORT} -f {function_name} -c \"{class_name}\" -p \"n:{param}\""
    # Esegui il comando
    os.system(command)

# Assumendo che il primo evento si verifichi a tempo 0
start_time = 0.0

# Iterare sulla lista degli arrivi
for i, (arrival_time, function_name, param, class_name) in enumerate(complete_list):
    if i == 0:
        delay = arrival_time - start_time
    else:
        delay = arrival_time - complete_list[i-1][0]
    
    time.sleep(delay)

    # Creare e avviare un thread per stampare il nome della funzione
    threading.Thread(target=invoke_function, args=(function_name, param, class_name)).start()
