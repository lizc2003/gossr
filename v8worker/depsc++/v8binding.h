// Copyright 2020-present, lizc2003@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

#ifndef BINDING_H
#define BINDING_H
#ifdef __cplusplus
extern "C" {
#endif

struct V8Worker_s;
typedef struct V8Worker_s V8Worker;

void         v8_init(char* icu_path);
const char*  v8_version();
V8Worker*    v8_worker_new(int table_index);
void         v8_worker_dispose(V8Worker* w);
const char*  v8_last_exception(V8Worker* w);
int          v8_execute(V8Worker* w, char* name, char* source);
int          v8_send(V8Worker* w, int type, char* s);
void         v8_terminate_execution(V8Worker* w);

extern void  go_v8SendCallback(int table_index, int type, char* msg, int msgLen, long long userdata);
extern char* go_v8RequestCallback(int table_index, int type, char* msg, int msgLen);

#ifdef __cplusplus
}  // extern "C"
#endif
#endif
