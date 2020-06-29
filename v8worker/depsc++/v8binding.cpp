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

#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <string>
#include <map>

#include "v8binding.h"
#include "libplatform/libplatform.h"
#include "v8.h"

using namespace v8;

struct V8Worker_s {
  int table_index;
  ArrayBuffer::Allocator* allocator;
  Isolate* isolate;
  std::string last_exception;
  Persistent<Function> recv_func;
  Persistent<Context> context;
};

void outputDebug(V8Worker* w, const char* msg) {
  //printf("%s\n", msg);
  go_v8SendCallback(w->table_index, 0, (char*)msg, strlen(msg), 0);
}

const char* toCString(const String::Utf8Value& value) {
  const char* s = *value;
  return s ? s : "<string conversion failed>";
}

std::string getExceptionString(V8Worker* w, TryCatch* try_catch) {
  Isolate* isolate = w->isolate;
  Local<Message> message = try_catch->Message();
  Local<Value> exception_obj = try_catch->Exception();
  std::string out;
  char tmpbuf[128];

  HandleScope handle_scope(isolate);
  Local<Context> context = isolate->GetCurrentContext();
  bool enter_context = context.IsEmpty();
  if (enter_context) {
    context = Local<Context>::New(isolate, w->context);
    context->Enter();
  }

  String::Utf8Value exception(isolate, exception_obj);
  const char* exception_string = toCString(exception);
  if (message.IsEmpty()) {
    // V8 didn't provide any extra information about this error; just
    // print the exception.
    out.append(exception_string);
    out.append("\n");
  } else if (message->GetScriptOrigin().Options().IsWasm()) {
    // Print wasm-function[(function index)]:(offset): (message).
    int function_index = message->GetWasmFunctionIndex();
    int offset = message->GetStartColumn(context).FromJust();
    sprintf(tmpbuf, "wasm-function[%d]:0x%x: ", function_index, offset);
    out.append(tmpbuf);
    out.append(exception_string);
    out.append("\n");
  } else {
    // Print (filename):(line number): (message).
    v8::String::Utf8Value filename(isolate,
                                   message->GetScriptOrigin().ResourceName());
    const char* filename_string = toCString(filename);
    int linenum = message->GetLineNumber(context).FromMaybe(-1);
    sprintf(tmpbuf, ":%i: ", linenum);
    out.append(filename_string);
    out.append(tmpbuf);
    out.append(exception_string);
    out.append("\n");
    Local<String> sourceline;
    if (message->GetSourceLine(context).ToLocal(&sourceline)) {
      // Print line of source code.
      v8::String::Utf8Value sourcelinevalue(isolate, sourceline);
      const char* sourceline_string = toCString(sourcelinevalue);
      out.append(sourceline_string);
      out.append("\n");
      // Print wavy underline (GetUnderline is deprecated).
      int start = message->GetStartColumn(context).FromJust();
      for (int i = 0; i < start; i++) {
        out.append(" ");
      }
      int end = message->GetEndColumn(context).FromJust();
      for (int i = start; i < end; i++) {
        out.append("^");
      }
      out.append("\n");
    }
  }
  if (enter_context) context->Exit();
  return out;
}

void v8implRequest(const FunctionCallbackInfo<Value>& args) {
  Isolate* isolate = args.GetIsolate();
  V8Worker* w = static_cast<V8Worker*>(isolate->GetData(0));
  assert(w->isolate == isolate);

  int argc = args.Length();
  if (argc < 2) {
    return;
  }

  int type = 0;
  std::string msg;
  {
    HandleScope handle_scope(isolate);
    Local<Context> context = isolate->GetCurrentContext();

    type = (int)args[0]->Int32Value(context).FromMaybe(0);

    String::Utf8Value str(isolate, args[1]);
    msg = toCString(str);
  }

  char* returnMsg = go_v8RequestCallback(w->table_index, type, (char*)msg.c_str(), (int)msg.size());
  Local<String> returnV = String::NewFromUtf8(isolate, returnMsg).ToLocalChecked();
  args.GetReturnValue().Set(returnV);
  free(returnMsg);
}

void v8implSend(const FunctionCallbackInfo<Value>& args) {
  Isolate* isolate = args.GetIsolate();
  V8Worker* w = static_cast<V8Worker*>(isolate->GetData(0));
  assert(w->isolate == isolate);

  int argc = args.Length();
  if (argc < 2) {
    return;
  }

  int type = 0;
  long long userdata = 0;
  std::string msg;
  {
    HandleScope handle_scope(isolate);
    Local<Context> context = isolate->GetCurrentContext();

    type = (int)args[0]->Int32Value(context).FromMaybe(0);

    String::Utf8Value str(isolate, args[1]);
    msg = toCString(str);

    if (argc > 2) {
      userdata = (long long)args[2]->IntegerValue(context).FromMaybe(0);
    }
  }
  go_v8SendCallback(w->table_index, type, (char*)msg.c_str(), (int)msg.size(), userdata);
}

void v8implPrint(const FunctionCallbackInfo<Value>& args) {
  Isolate* isolate = args.GetIsolate();
  V8Worker* w = static_cast<V8Worker*>(isolate->GetData(0));
  assert(w->isolate == isolate);

  int argc = args.Length();
  if (argc < 2) {
    return;
  }

  Local<Context> context = isolate->GetCurrentContext();

  int type = (int)args[0]->Int32Value(context).FromMaybe(0);

  std::string out;
  bool first = true;
  for (int i = 1; i < args.Length(); i++) {
    HandleScope handle_scope(isolate);

    TryCatch try_catch(isolate);
    if (first) {
      first = false;
    } else {
      out += " ";
    }
    Local<Value> arg = args[i];
    if (arg->IsObject()) {
      Local<Object> obj = arg->ToObject(context).ToLocalChecked();
      Local<Array> props = obj->GetOwnPropertyNames(context).ToLocalChecked();
      int len = props->Length();
      if (len == 0) {
        out += toCString(String::Utf8Value(isolate, arg));
      } else {
        out += "{";
        bool first2 = true;
        for(int i = 0, l = len; i < l; i++) {
          if (first2) {
            first2 = false;
          } else {
            out += ",";
          }
          Local<Value> localKey = props->Get(context, i).ToLocalChecked();
          Local<Value> localVal = obj->Get(context, localKey).ToLocalChecked();
          out += toCString(String::Utf8Value(isolate, localKey));
          out += ":";
          out += toCString(String::Utf8Value(isolate, localVal));
        }
        out += "}";
      }
    } else {
      out += toCString(String::Utf8Value(isolate, arg));
    }
  }
  go_v8SendCallback(w->table_index, type, (char*)out.c_str(), (int)out.size(), 0);
}

void v8implSetRecv(const FunctionCallbackInfo<Value>& args) {
  Isolate* isolate = args.GetIsolate();
  V8Worker* w = (V8Worker*)isolate->GetData(0);
  assert(w->isolate == isolate);

  HandleScope handle_scope(isolate);
  Local<Function> func = Local<Function>::Cast(args[0]);
  if (!func.IsEmpty()) {
    w->recv_func.Reset(isolate, func);
  }
}

extern "C" {

static std::unique_ptr<v8::Platform> g_v8platform;

const char* v8_version() { return V8::GetVersion(); }

const char* v8_last_exception(V8Worker* w) {
  return w->last_exception.c_str();
}

void v8_init(char* icu_path) {
  if (g_v8platform.get() == nullptr) {
    V8::InitializeICU(icu_path);
    g_v8platform = platform::NewDefaultPlatform();
    V8::InitializePlatform(g_v8platform.get());
    V8::Initialize();
    // TODO(ry) This makes WASM compile synchronously. Eventually we should
    // remove this to make it work asynchronously too. But that requires getting
    // PumpMessageLoop and RunMicrotasks setup correctly.
    // See https://github.com/denoland/deno/issues/2544
    const char* argv[3] = {"", "--no-wasm-async-compilation", "--max-old-space-size=4096"};
    int argc = 3;
    V8::SetFlagsFromCommandLine(&argc, const_cast<char**>(argv), false);
  }
}

V8Worker* v8_worker_new(int table_index) {
  V8Worker* w = new V8Worker;

  Isolate::CreateParams create_params;
  w->allocator = ArrayBuffer::Allocator::NewDefaultAllocator();
  create_params.array_buffer_allocator = w->allocator;

  Isolate* isolate = Isolate::New(create_params);
  w->isolate = isolate;
  w->isolate->SetData(0, w);
  w->table_index = table_index;

  Locker locker(isolate);
  Isolate::Scope isolate_scope(isolate);
  HandleScope handle_scope(isolate);

  // Create a template for the global object.
  Local<ObjectTemplate> global = ObjectTemplate::New(isolate);
  Local<ObjectTemplate> v8worker = ObjectTemplate::New(isolate);

  global->Set(String::NewFromUtf8(isolate, "v8worker").ToLocalChecked(), v8worker);

  v8worker->Set(String::NewFromUtf8(isolate, "print").ToLocalChecked(),
                 FunctionTemplate::New(isolate, v8implPrint));

  v8worker->Set(String::NewFromUtf8(isolate, "setRecv").ToLocalChecked(),
                 FunctionTemplate::New(isolate, v8implSetRecv));

  v8worker->Set(String::NewFromUtf8(isolate, "send").ToLocalChecked(),
                 FunctionTemplate::New(isolate, v8implSend));

  v8worker->Set(String::NewFromUtf8(isolate, "request").ToLocalChecked(),
                 FunctionTemplate::New(isolate, v8implRequest));

  Local<Context> context = Context::New(isolate, NULL, global);
  w->context.Reset(isolate, context);
  //context->Enter();

  //outputDebug(w, "=========== worker new");
  return w;
}

void v8_worker_dispose(V8Worker* w) {
  w->recv_func.Reset();
  w->context.Reset();
  w->isolate->Dispose();
  delete w->allocator;
  delete (w);
}

void v8_terminate_execution(V8Worker* w) {
  w->isolate->TerminateExecution();
}

int v8_execute(V8Worker* w, char* name_s, char* source_s) {
  Isolate* isolate = w->isolate;
  Locker locker(isolate);
  Isolate::Scope isolate_scope(isolate);
  // Create a stack-allocated handle scope.
  HandleScope handle_scope(isolate);

  Local<Context> context = Local<Context>::New(isolate, w->context);
  Context::Scope context_scope(context);

  TryCatch try_catch(isolate);

  Local<String> name = String::NewFromUtf8(isolate, name_s).ToLocalChecked();
  Local<String> source = String::NewFromUtf8(isolate, source_s).ToLocalChecked();

  ScriptOrigin origin(name);
  MaybeLocal<Script> script = Script::Compile(context, source, &origin);
  if (script.IsEmpty()) {
    w->last_exception = getExceptionString(w, &try_catch);
    return 1;
  }

  MaybeLocal<Value> result = script.ToLocalChecked()->Run(context);
  if (result.IsEmpty()) {
    w->last_exception = getExceptionString(w, &try_catch);
    return 2;
  }

  return 0;
}

int v8_send(V8Worker* w, int type, char* s) {
  Isolate* isolate = w->isolate;
  Locker locker(isolate);
  Isolate::Scope isolate_scope(isolate);
  HandleScope handle_scope(isolate);

  Local<Context> context = Local<Context>::New(isolate, w->context);

  Local<Function> recvFunc = Local<Function>::New(isolate, w->recv_func);
  if (recvFunc.IsEmpty()) {
    w->last_exception = "v8worker.recv has not been called.";
    return 1;
  }

  Local<Value> args[2];
  args[0] = Integer::New(isolate, type);
  args[1] = String::NewFromUtf8(isolate, s).ToLocalChecked();

  TryCatch try_catch(isolate);
  MaybeLocal<Value> result = recvFunc->Call(context, Undefined(isolate), 2, args);
  result.IsEmpty();

  if (try_catch.HasCaught()) {
    w->last_exception = getExceptionString(w, &try_catch);
    return 2;
  }

  return 0;
}

}
