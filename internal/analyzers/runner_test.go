package analyzers

import (
 "os"
 "path/filepath"
 "testing"
)

func TestExternalScannerBinary(t *testing.T) {
 base:=t.TempDir()
 root:=filepath.Join(base,"assessed")
 if err:=os.Mkdir(root,0700);err!=nil{t.Fatal(err)}
 root,err:=canonicalRoot(root);if err!=nil{t.Fatal(err)}
 external:=filepath.Join(base,"scanner")
 if err:=os.WriteFile(external,nil,0700);err!=nil{t.Fatal(err)}
 if err:=externalScannerBinary(root,external);err!=nil{t.Fatalf("external scanner rejected: %v",err)}
 internal:=filepath.Join(root,"scanner")
 if err:=os.WriteFile(internal,nil,0700);err!=nil{t.Fatal(err)}
 if err:=externalScannerBinary(root,internal);err==nil{t.Fatal("assessed repository scanner accepted")}
}
