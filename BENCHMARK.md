# HTTP Benchmark Results for `/get` Endpoint

All tests were run with [hey](https://github.com/rakyll/hey) against `http://localhost:8080/get` with **10,000 total requests** and varying concurrency levels.

---

## 1. 50 Concurrent Requests

**Command:**  
```sh
hey -n 10000 -c 50 http://localhost:8080/get
```

**Summary:**

| Metric           | Value        |
|------------------|-------------|
| Total Time       | 2.6422 sec  |
| Slowest          | 0.0277 sec  |
| Fastest          | 0.0009 sec  |
| Average          | 0.0131 sec  |
| Requests/sec     | 3784.79     |
| Total Data       | 1,419,957 B |
| Size/Request     | 141 B       |
| Status Codes     | [200] 10,000|

**Latency distribution:**

| Percentile | Value (sec) |
|------------|------------|
| 10%        | 0.0093     |
| 25%        | 0.0108     |
| 50%        | 0.0131     |
| 75%        | 0.0152     |
| 90%        | 0.0169     |
| 95%        | 0.0179     |
| 99%        | 0.0198     |

**Response time histogram:**
```
  0.001 [1]     |
  0.004 [13]    |
  0.006 [52]    |■
  0.009 [689]   |■■■■■■■■
  0.012 [2363]  |■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.014 [3520]  |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.017 [2406]  |■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.020 [844]   |■■■■■■■■■■
  0.022 [83]    |■
  0.025 [17]    |
  0.028 [12]    |
```

---

## 2. 500 Concurrent Requests

**Command:**  
```sh
hey -n 10000 -c 500 http://localhost:8080/get
```

**Summary:**

| Metric           | Value        |
|------------------|-------------|
| Total Time       | 2.6694 sec  |
| Slowest          | 1.4924 sec  |
| Fastest          | 0.0008 sec  |
| Average          | 0.1053 sec  |
| Requests/sec     | 3746.17     |
| Total Data       | 1,483,944 B |
| Size/Request     | 148 B       |
| Status Codes     | [200] 10,000|

**Latency distribution:**

| Percentile | Value (sec) |
|------------|------------|
| 10%        | 0.0538     |
| 25%        | 0.0607     |
| 50%        | 0.0725     |
| 75%        | 0.0956     |
| 90%        | 0.1283     |
| 95%        | 0.1357     |
| 99%        | 1.1533     |

**Response time histogram:**
```
  0.001 [1]     |
  0.150 [9722]  |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.299 [46]    |
  0.448 [0]     |
  0.597 [0]     |
  0.747 [0]     |
  0.896 [0]     |
  1.045 [0]     |
  1.194 [145]   |■
  1.343 [51]    |
  1.492 [35]    |
```

---

## 3. 1000 Concurrent Requests

**Command:**  
```sh
hey -n 10000 -c 1000 http://localhost:8080/get
```

**Summary:**

| Metric           | Value        |
|------------------|-------------|
| Total Time       | 2.7662 sec  |
| Slowest          | 1.7863 sec  |
| Fastest          | 0.0009 sec  |
| Average          | 0.1974 sec  |
| Requests/sec     | 3615.01     |
| Total Data       | 1,387,617 B |
| Size/Request     | 138 B       |
| Status Codes     | [200] 10,000|

**Latency distribution:**

| Percentile | Value (sec) |
|------------|------------|
| 10%        | 0.0850     |
| 25%        | 0.1063     |
| 50%        | 0.1208     |
| 75%        | 0.1522     |
| 90%        | 0.1738     |
| 95%        | 1.2074     |
| 99%        | 1.4937     |

**Response time histogram:**
```
  0.001 [1]     |
  0.179 [9031]  |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.358 [383]   |■■
  0.537 [0]     |
  0.715 [0]     |
  0.894 [0]     |
  1.072 [0]     |
  1.251 [173]   |■
  1.429 [215]   |■
  1.608 [114]   |■
  1.786 [83]    |
```

---

## 4. 5000 Concurrent Requests

**Command:**  
```sh
hey -n 10000 -c 5000 http://localhost:8080/get
```

**Summary:**

| Metric           | Value        |
|------------------|-------------|
| Total Time       | 15.3971 sec |
| Slowest          | 15.3729 sec |
| Fastest          | 0.0029 sec  |
| Average          | 1.1998 sec  |
| Requests/sec     | 649.47      |
| Total Data       | 1,127,511 B |
| Size/Request     | 112 B       |
| Status Codes     | [200] 10,000|

**Latency distribution:**

| Percentile | Value (sec) |
|------------|------------|
| 10%        | 0.0374     |
| 25%        | 0.1400     |
| 50%        | 0.4508     |
| 75%        | 1.6893     |
| 90%        | 3.3092     |
| 95%        | 4.0002     |
| 99%        | 8.4167     |

**Response time histogram:**
```
  0.003 [1]     |
  1.540 [7092]  |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  3.077 [1781]  |■■■■■■■■■■
  4.614 [733]   |■■■■
  6.151 [128]   |■
  7.688 [95]    |■
  9.225 [147]   |■
  10.762 [0]    |
  12.299 [0]    |
  13.836 [0]    |
  15.373 [23]   |
```

---

## Observations

- As concurrency increases, average latency and slowest requests rise significantly.
- Throughput (requests/sec) drops sharply at the highest concurrency (5000).
- All requests returned 200 status codes for every test.

## Docker Resources for Benchmark

Given the resources of the host machine, the following resources were allocated to the container:
- CPU: 1 core
- Memory: 256MB
- Max Open Files: 100000

---

*Tested with hey on: 2025-05-22*
