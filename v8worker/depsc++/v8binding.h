#ifndef BINDING_H
#define BINDING_H
#ifdef __cplusplus
extern "C" {
#endif

struct V8Worker_s;
typedef struct V8Worker_s V8Worker;

void         v8_init();
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
