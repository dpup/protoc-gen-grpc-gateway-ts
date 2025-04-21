package registry

import (
	"github.com/dpup/protoc-gen-grpc-gateway-ts/data"
	"google.golang.org/protobuf/types/descriptorpb"
)

// getFieldType generates an intermediate type and leave the rendering logic to
// choose what to render.
func (r *Registry) getFieldType(f *descriptorpb.FieldDescriptorProto) string {
	if f.Type == nil {
		return ""
	}
	switch *f.Type {
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE,
		descriptorpb.FieldDescriptorProto_TYPE_ENUM,
		descriptorpb.FieldDescriptorProto_TYPE_GROUP:
		return f.GetTypeName()
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return "string"
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return "bool"
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return "bytes"
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		return "float"
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		return "double"
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		return "fixed32"
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		return "sfixed32"
	case descriptorpb.FieldDescriptorProto_TYPE_INT32:
		return "int32"
	case descriptorpb.FieldDescriptorProto_TYPE_SINT32:
		return "sint32"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32:
		return "uint32"
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		return "fixed64"
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		return "sfixed64"
	case descriptorpb.FieldDescriptorProto_TYPE_INT64:
		return "int64"
	case descriptorpb.FieldDescriptorProto_TYPE_SINT64:
		return "sint64"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64:
		return "uint64"
	default:
		return ""
	}
}

func (r *Registry) analyseField(
	fileData *data.File,
	msgData *data.Message,
	packageName string,
	f *descriptorpb.FieldDescriptorProto) {
	fqTypeName := r.getFieldType(f)
	isExternal := r.isExternalDependenciesOutsidePackage(fqTypeName, packageName)

	fieldData := &data.Field{
		Name:         f.GetName(),
		Type:         fqTypeName,
		IsExternal:   isExternal,
		IsOneOfField: f.OneofIndex != nil,
		IsOptional:   f.GetProto3Optional(),
		IsDeprecated: f.GetOptions().GetDeprecated(),
		JsonName:     f.GetJsonName(),
		Message:      msgData,
	}

	if f.Label != nil {
		if f.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
			fieldData.IsRepeated = true
		}
	}

	msgData.Fields = append(msgData.Fields, fieldData)

	if !fieldData.IsOneOfField {
		msgData.NonOneOfFields = append(msgData.NonOneOfFields, fieldData)
	}
	if fieldData.IsOptional {
		msgData.OptionalFields = append(msgData.OptionalFields, fieldData)
	}

	// if it's an external dependencies. store in the file data so that they can be collected when every file's finished
	if isExternal {
		fileData.ExternalDependingTypes = append(fileData.ExternalDependingTypes, fqTypeName)
	}

	// if it's a one of field. register the field data in the group of the same one of index.
	// internally, optional fields are modeled as OneOf, however, we don't want to include them here.
	if fieldData.IsOneOfField && !fieldData.IsOptional {
		index := f.GetOneofIndex()
		fieldData.OneOfIndex = index
		_, ok := msgData.OneOfFieldsGroups[index]
		if !ok {
			msgData.OneOfFieldsGroups[index] = make([]*data.Field, 0)
		}
		msgData.OneOfFieldsGroups[index] = append(msgData.OneOfFieldsGroups[index], fieldData)
	}

	fileData.TrackPackageNonScalarType(fieldData)
}
