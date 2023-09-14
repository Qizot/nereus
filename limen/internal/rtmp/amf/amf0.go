package amf

var (
  NumberType = byte(0x00)
  BooleanType = byte(0x01)
  StringType = byte(0x02)
  KeyValueObjectType = byte(0x03)
  NullType = byte(0x05)
  ECMAArrayType = byte(0x08)
  ObjectEndMarker = [3]byte{0x00, 0x00, 0x09}
)
