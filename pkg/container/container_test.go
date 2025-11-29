package container

import (
	"errors"
	"testing"
)

func TestContainer_SetAndResolve(t *testing.T) {
	Reset() // 确保测试环境干净

	// Test: 设置和解析简单值
	App().Set("test_string", "hello")
	
	value, err := App().Resolve("test_string")
	if err != nil {
		t.Fatalf("Failed to resolve: %v", err)
	}
	
	str, ok := value.(string)
	if !ok {
		t.Fatal("Value is not a string")
	}
	
	if str != "hello" {
		t.Fatalf("Expected 'hello', got '%s'", str)
	}
}

func TestContainer_Bind(t *testing.T) {
	Reset()

	// Test: 使用工厂函数绑定
	called := 0
	App().Bind("factory_test", func() (any, error) {
		called++
		return "factory_value", nil
	})
	
	// 第一次解析应该调用工厂函数
	value1, err := App().Resolve("factory_test")
	if err != nil {
		t.Fatalf("Failed to resolve: %v", err)
	}
	
	if called != 1 {
		t.Fatalf("Expected factory to be called once, called %d times", called)
	}
	
	// 第二次解析应该返回缓存的值
	value2, err := App().Resolve("factory_test")
	if err != nil {
		t.Fatalf("Failed to resolve: %v", err)
	}
	
	if called != 1 {
		t.Fatalf("Expected factory to be called once (singleton), called %d times", called)
	}
	
	if value1 != value2 {
		t.Fatal("Expected same instance (singleton)")
	}
}

func TestContainer_BindError(t *testing.T) {
	Reset()

	// Test: 工厂函数返回错误
	testErr := errors.New("factory error")
	App().Bind("error_test", func() (any, error) {
		return nil, testErr
	})
	
	_, err := App().Resolve("error_test")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	
	if err != testErr {
		t.Fatalf("Expected %v, got %v", testErr, err)
	}
}

func TestContainer_Has(t *testing.T) {
	Reset()

	// Test: 检查服务是否存在
	if App().Has("non_existent") {
		t.Fatal("Expected false for non-existent service")
	}
	
	App().Set("exists", "value")
	if !App().Has("exists") {
		t.Fatal("Expected true for existing service")
	}
	
	App().Bind("factory_exists", func() (any, error) {
		return "value", nil
	})
	if !App().Has("factory_exists") {
		t.Fatal("Expected true for bound factory")
	}
}

func TestContainer_MustResolve(t *testing.T) {
	Reset()

	// Test: MustResolve 成功情况
	App().Set("test", "value")
	
	defer func() {
		if r := recover(); r != nil {
			t.Fatal("MustResolve should not panic for existing service")
		}
	}()
	
	value := App().MustResolve("test")
	if value != "value" {
		t.Fatalf("Expected 'value', got %v", value)
	}
}

func TestContainer_MustResolvePanic(t *testing.T) {
	Reset()

	// Test: MustResolve panic 情况
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustResolve should panic for non-existent service")
		}
	}()
	
	App().MustResolve("non_existent")
}

func TestResolveAs(t *testing.T) {
	Reset()

	// Test: 类型安全的解析
	App().Set("string_value", "hello")
	
	str, err := ResolveAs[string]("string_value")
	if err != nil {
		t.Fatalf("Failed to resolve: %v", err)
	}
	
	if str != "hello" {
		t.Fatalf("Expected 'hello', got '%s'", str)
	}
}

func TestResolveAs_TypeMismatch(t *testing.T) {
	Reset()

	// Test: 类型不匹配
	App().Set("string_value", "hello")
	
	_, err := ResolveAs[int]("string_value")
	if err == nil {
		t.Fatal("Expected error for type mismatch")
	}
}

func TestMustResolveAs(t *testing.T) {
	Reset()

	// Test: MustResolveAs 成功情况
	App().Set("int_value", 42)
	
	defer func() {
		if r := recover(); r != nil {
			t.Fatal("MustResolveAs should not panic for correct type")
		}
	}()
	
	value := MustResolveAs[int]("int_value")
	if value != 42 {
		t.Fatalf("Expected 42, got %d", value)
	}
}

func TestMustResolveAs_Panic(t *testing.T) {
	Reset()

	// Test: MustResolveAs panic 情况
	App().Set("string_value", "hello")
	
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustResolveAs should panic for type mismatch")
		}
	}()
	
	MustResolveAs[int]("string_value")
}

func TestContainer_Keys(t *testing.T) {
	Reset()

	// Test: 获取所有键
	App().Set("key1", "value1")
	App().Set("key2", "value2")
	App().Bind("key3", func() (any, error) {
		return "value3", nil
	})
	
	keys := App().Keys()
	if len(keys) != 3 {
		t.Fatalf("Expected 3 keys, got %d", len(keys))
	}
	
	// 检查所有键是否存在
	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k] = true
	}
	
	if !keyMap["key1"] || !keyMap["key2"] || !keyMap["key3"] {
		t.Fatal("Missing expected keys")
	}
}

func TestContainer_SetOverridesFactory(t *testing.T) {
	Reset()

	// Test: Set 应该覆盖已绑定的工厂
	App().Bind("test", func() (any, error) {
		return "factory", nil
	})
	
	App().Set("test", "direct")
	
	value, err := App().Resolve("test")
	if err != nil {
		t.Fatalf("Failed to resolve: %v", err)
	}
	
	if value != "direct" {
		t.Fatalf("Expected 'direct', got %v", value)
	}
}

func TestContainer_BindOverridesSet(t *testing.T) {
	Reset()

	// Test: Bind 应该覆盖已设置的值
	App().Set("test", "direct")
	
	App().Bind("test", func() (any, error) {
		return "factory", nil
	})
	
	value, err := App().Resolve("test")
	if err != nil {
		t.Fatalf("Failed to resolve: %v", err)
	}
	
	if value != "factory" {
		t.Fatalf("Expected 'factory', got %v", value)
	}
}

// Benchmark tests
func BenchmarkContainer_Set(b *testing.B) {
	Reset()
	for i := 0; i < b.N; i++ {
		App().Set("bench_test", "value")
	}
}

func BenchmarkContainer_Resolve(b *testing.B) {
	Reset()
	App().Set("bench_test", "value")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = App().Resolve("bench_test")
	}
}

func BenchmarkContainer_ResolveAs(b *testing.B) {
	Reset()
	App().Set("bench_test", "value")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ResolveAs[string]("bench_test")
	}
}
